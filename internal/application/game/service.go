package game

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
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
}

// killSignal reports the outcome of a verified kill attempt: either a real
// kill, or a reap when the target's process had already exited.
type killSignal struct {
	target     *game.Target
	shouldReap bool
}

// scoreTracker accumulates kills and freed memory for a single live session,
// tracking the running high score seeded from the caller's persisted best.
type scoreTracker struct {
	kills     int
	freedMem  int64
	highScore int
}

// PlayResult carries the outcome of a completed play session, needed by the
// caller to record a score.
type PlayResult struct {
	Duration    float64
	LowestSpeed float64
	Kills       int
	FreedMem    int64
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

// Play runs the game loop for the given already-discovered processes.
// highScore is the caller's persisted best, used to track a running high
// score for display during the session.
func (s *Service) Play(cfg inbound.Config, processes []process.Info, highScore int) (PlayResult, error) {
	session := game.NewSession(processes, game.Config{Confirm: cfg.ConfirmMode, Speed: cfg.Speed, TimeLimit: cfg.TimeLimit})

	tracker := &scoreTracker{highScore: highScore}
	endTime, err := s.runLoop(session, tracker)
	if err != nil {
		return PlayResult{}, err
	}

	return PlayResult{
		Duration:    endTime.Sub(session.StartTime()).Seconds(),
		LowestSpeed: session.Throttle().LowestSpeed(),
		Kills:       tracker.kills,
		FreedMem:    tracker.freedMem,
	}, nil
}

func (s *Service) runLoop(session *game.Session, tracker *scoreTracker) (time.Time, error) {
	if err := s.renderer.Init(); err != nil {
		return time.Time{}, fmt.Errorf("renderer initialization failed: %w", err)
	}
	defer s.renderer.Cleanup()

	session.Start(s.renderer.Size())

	termSignal, stopSignals := s.registerTermSignalWatcher(session)
	defer stopSignals()

	// Closing done unblocks any killOrReap goroutine waiting to send on
	// killSignals after frameLoop exits, preventing a goroutine leak.
	done := make(chan struct{})
	defer close(done)

	dispatcher := event.NewDispatcher(game.NewInput(session))

	s.frameLoop(session, tracker, dispatcher, termSignal, done)

	// Capture end time before deferred cleanup runs.
	return time.Now(), nil
}

// registerTermSignalWatcher stops game session when the process receives an interrupt,
// termination, or suspend signal. termSignal is closed when that happens, letting
// callers react immediately instead of waiting out a polling interval. The
// returned stop func deregisters the watcher and must be called once the
// caller is done with game session.
func (s *Service) registerTermSignalWatcher(session *game.Session) (termSignal <-chan struct{}, stop func()) {
	// ctx.Done() fires on either a real signal or an explicit stop() call;
	// session.Stop() is idempotent, so it's harmless to call it in both cases.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	go func() {
		<-ctx.Done()
		session.Stop()
	}()
	return ctx.Done(), stop
}

// frameLoop drives the game at frameDuration cadence until session stops running.
// It wakes immediately when session stops mid-frame or termSignal fires, instead of
// waiting out the remainder of the current tick.
func (s *Service) frameLoop(session *game.Session, tracker *scoreTracker, dispatcher *event.Dispatcher, termSignal <-chan struct{}, done <-chan struct{}) {
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	killSignals := make(chan killSignal, 10)

	for session.IsRunning() {
		s.applyKillSignals(tracker, killSignals)
		s.drainEventQueue(dispatcher, killSignals, done)
		session.Update(s.renderer.Size())
		s.renderer.Render(toFrameState(session, tracker))

		if !session.IsRunning() {
			break
		}

		select {
		case <-ticker.C:
		case <-termSignal:
		}
	}
}

func (s *Service) applyKillSignals(tracker *scoreTracker, killSignals <-chan killSignal) {
	for {
		select {
		case sig := <-killSignals:
			switch {
			case sig.shouldReap:
				sig.target.Reap()
			case sig.target.Kill():
				tracker.recordKill(sig.target.Rss)
			}
		default:
			return
		}
	}
}

func (s *Service) drainEventQueue(dispatcher *event.Dispatcher, kills chan<- killSignal, done <-chan struct{}) {
	events := s.events.Events()
	for {
		select {
		case inputEvent := <-events:
			if target := dispatcher.Dispatch(inputEvent); target != nil {
				go s.killOrReap(target, kills, done)
			}
		default:
			return
		}
	}
}

func (s *Service) killOrReap(target *game.Target, kills chan<- killSignal, done <-chan struct{}) {
	killed, shouldReap, err := s.killer.Kill(target)

	sig := killSignal{target: target}
	if err != nil || !killed {
		if !shouldReap {
			return
		}
		sig.shouldReap = true
	}

	select {
	case kills <- sig:
	case <-done:
	}
}

// recordKill records a kill, updates freed memory, and raises the high score when needed.
func (t *scoreTracker) recordKill(freedMemory int64) {
	t.kills++
	t.freedMem += freedMemory
	if t.kills > t.highScore {
		t.highScore = t.kills
	}
}
