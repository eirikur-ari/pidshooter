package game

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

const frameDuration = time.Second / 20

// killer verifies and terminates the target's backing OS process, reporting
// whether the caller should reap the target because its process was already gone.
type killer interface {
	Kill(target *game.Target) (killed, shouldReap bool, err error)
}

// Service orchestrates core domain objects and outbound ports to play a single game session.
type Service struct {
	killer   killer
	renderer outbound.Renderer
	events   outbound.InputSource
	kills    chan killSignal
}

// killSignal reports the outcome of a verified kill attempt: either a real
// kill, or a reap when the target's process had already exited.
type killSignal struct {
	target     *game.Target
	shouldReap bool
}

// NewService constructs a Service with all required outbound ports injected.
func NewService(
	killer killer,
	renderer outbound.Renderer,
	events outbound.InputSource,
) *Service {
	return &Service{
		killer:   killer,
		renderer: renderer,
		events:   events,
	}
}

// PlayResult carries the outcome of a completed play session, needed by the
// caller to record a score.
type PlayResult struct {
	Duration    float64
	LowestSpeed float64
}

// Play runs the game loop for the given already-discovered processes, recording kills against tracker.
func (s *Service) Play(cfg inbound.Config, processes []process.Info, tracker *score.Tracker) (PlayResult, error) {
	gs := s.newGame(cfg, processes)

	endTime, err := s.runLoop(gs, tracker)
	if err != nil {
		return PlayResult{}, err
	}

	return PlayResult{
		Duration:    endTime.Sub(gs.StartTime()).Seconds(),
		LowestSpeed: gs.Throttle().LowestSpeed(),
	}, nil
}

func (s *Service) newGame(cfg inbound.Config, processes []process.Info) *game.Session {
	return game.NewSession(processes, game.Config{Confirm: cfg.ConfirmMode, Speed: cfg.Speed, TimeLimit: cfg.TimeLimit})
}

func (s *Service) runLoop(gs *game.Session, tracker *score.Tracker) (time.Time, error) {
	if err := s.renderer.Init(); err != nil {
		return time.Time{}, fmt.Errorf("renderer initialization failed: %w", err)
	}
	defer s.renderer.Cleanup()

	gs.Start(s.renderer.Size())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	defer signal.Stop(sigCh)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-sigCh:
			gs.Stop()
		case <-done:
		}
	}()

	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	s.kills = make(chan killSignal, 10)
	evt := event.NewDispatcher(game.NewInput(gs))

	for gs.IsRunning() {
		s.applyKills(tracker)
		s.drainEvents(evt, done)
		gs.Update(s.renderer.Size())
		s.renderer.Render(buildFrame(gs, tracker))
		<-ticker.C
	}
	// Capture end time before deferred Cleanup() runs.
	return time.Now(), nil
}

func (s *Service) applyKills(tracker *score.Tracker) {
	for {
		select {
		case sig := <-s.kills:
			if sig.shouldReap {
				sig.target.Reap()
				continue
			}
			if sig.target.Kill() {
				tracker.RecordKill(sig.target.Rss)
			}
		default:
			return
		}
	}
}

func buildFrame(gs *game.Session, tracker *score.Tracker) outbound.FrameState {
	targets, alive := gs.AvailableTargets()

	return outbound.FrameState{
		Targets:   toTargetViewStates(targets),
		HUD:       toHUDState(tracker),
		StatusBar: toStatusState(gs, alive),
	}
}

func (s *Service) drainEvents(d *event.Dispatcher, done <-chan struct{}) {
	for {
		select {
		case ev := <-s.events.Events():
			target := d.Dispatch(ev)
			if target == nil {
				continue
			}
			go func() {
				killed, shouldReap, err := s.killer.Kill(target)
				switch {
				case err == nil && killed:
					select {
					case s.kills <- killSignal{target: target}:
					case <-done:
					}
				case shouldReap:
					select {
					case s.kills <- killSignal{target: target, shouldReap: true}:
					case <-done:
					}
				}
			}()
		default:
			return
		}
	}
}
