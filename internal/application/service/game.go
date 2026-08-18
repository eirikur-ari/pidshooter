package service

import (
	"errors"
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

// GameService implements inbound.GamePlay by orchestrating core domain objects and outbound ports.
type GameService struct {
	process  outbound.Process
	scores   *ScoreService
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

// errAlreadyKilled indicates a target's backing OS process was already gone
// by the time its kill could be verified.
var errAlreadyKilled = errors.New("target process already killed")

// NewGameService constructs a GameService with all required outbound ports injected.
func NewGameService(
	process outbound.Process,
	store outbound.ScoreStore,
	renderer outbound.Renderer,
	events outbound.InputSource,
) *GameService {
	return &GameService{
		process:  process,
		scores:   NewScoreService(store),
		renderer: renderer,
		events:   events,
	}
}

// Play runs a complete game session: discovery → game loop → score persistence → display.
func (s *GameService) Play(cfg inbound.GamePlayConfig) error {
	processes, err := s.findProcesses(cfg.Patterns)
	if err != nil {
		return err
	}

	if len(processes) == 0 {
		return nil
	}

	board, tracker, success := s.scores.loadScoreBoard()

	g := s.newGame(cfg, processes)

	endTime, err := s.runLoop(g, tracker)
	if err != nil {
		return err
	}

	duration := endTime.Sub(g.StartTime()).Seconds()

	s.scores.recordScore(board, g.Throttle().LowestSpeed(), cfg.TimeLimit, duration, success)

	printResults(duration, board)

	return nil
}

func (s *GameService) findProcesses(patterns []string) ([]process.Info, error) {
	if err := validateSearchPatterns(patterns); err != nil {
		return nil, err
	}
	processes, err := s.process.List()
	if err != nil {
		return nil, fmt.Errorf("process search failed: %w", err)
	}

	matches := process.Find(toProcessInfos(processes), patterns, s.process.OwnPid())

	if len(matches) == 0 {
		fmt.Printf("No processes found matching %v\n", patterns)
		return nil, nil
	}

	fmt.Printf("Found %d process(es) matching %v. Starting game...\n", len(matches), patterns)

	return matches, nil
}

func (s *GameService) newGame(cfg inbound.GamePlayConfig, processes []process.Info) *game.Game {
	return game.New(processes, game.Config{Confirm: cfg.ConfirmMode, Speed: cfg.Speed, TimeLimit: cfg.TimeLimit})
}

func (s *GameService) runLoop(g *game.Game, tracker *score.Tracker) (time.Time, error) {
	if err := s.renderer.Init(); err != nil {
		return time.Time{}, fmt.Errorf("renderer initialization failed: %w", err)
	}
	defer s.renderer.Cleanup()

	w, h := s.renderer.Size()
	g.Start(w, h)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	defer signal.Stop(sigCh)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-sigCh:
			g.Stop()
		case <-done:
		}
	}()

	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	s.kills = make(chan killSignal, 10)
	evt := event.NewDispatcher(game.NewInput(g))

	for g.IsRunning() {
		s.applyKills(tracker)
		s.drainEvents(evt, done)
		w, h = s.renderer.Size()
		g.Update(w, h)
		s.renderer.Render(buildFrame(g, tracker))
		<-ticker.C
	}
	// Capture end time before deferred Cleanup() runs.
	return time.Now(), nil
}

func (s *GameService) applyKills(tracker *score.Tracker) {
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

func buildFrame(g *game.Game, tracker *score.Tracker) outbound.FrameState {
	frame := g.Frame()
	views := make([]outbound.TargetViewState, len(frame.Targets))
	for i, s := range frame.Targets {
		views[i] = outbound.TargetViewState{X: s.X, Y: s.Y, Tag: s.Tag, Killing: s.Killing}
	}

	return outbound.FrameState{
		Targets: views,
		HUD: outbound.HUDState{
			FreedMem:  tracker.FreedMem,
			Kills:     tracker.Kills,
			HighScore: tracker.HighScore,
		},
		StatusBar: outbound.StatusState{
			Alive:      frame.Alive,
			Speed:      g.Throttle().Speed(),
			TimeLimit:  g.TimeLimit(),
			TimeLeft:   g.TimeLeft(),
			Confirming: toConfirmViewState(g.ConfirmTarget()),
		},
	}
}

func toConfirmViewState(t *game.Target) *outbound.ConfirmViewState {
	if t == nil {
		return nil
	}
	return &outbound.ConfirmViewState{PID: t.Pid, Name: t.Name}
}

func (s *GameService) drainEvents(d *event.Dispatcher, done <-chan struct{}) {
	for {
		select {
		case ev := <-s.events.Events():
			target := d.Dispatch(ev)
			if target == nil {
				continue
			}
			go func() {
				killed, err := s.kill(target)
				switch {
				case err == nil && killed:
					select {
					case s.kills <- killSignal{target: target}:
					case <-done:
					}
				case errors.Is(err, errAlreadyKilled):
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

// kill re-verifies target immediately before signaling it, since the PID may
// have been recycled by the OS to a different process — or exited entirely —
// in the time between discovery and the player confirming the kill. Either
// case is reported as errAlreadyKilled so the caller can reap the target
// instead of leaving it stuck as Alive forever.
func (s *GameService) kill(target *game.Target) (bool, error) {
	pid := target.Pid
	if target.Info.IsProtected() {
		return false, fmt.Errorf("refusing to kill PID %d", pid)
	}
	name, err := s.process.LookupName(pid)
	if err != nil {
		return false, fmt.Errorf("could not verify PID %d: %w: %w", pid, errAlreadyKilled, err)
	}
	if err := validateProcessName(target.Name, name); err != nil {
		return false, fmt.Errorf("%w: %w", errAlreadyKilled, err)
	}

	return s.process.Kill(pid)
}
