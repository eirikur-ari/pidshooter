package game

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

const frameDuration = time.Second / 20

// Service orchestrates core domain objects and outbound ports to play a single game session.
type Service struct {
	killer   processKiller
	renderer outbound.Renderer
	events   outbound.InputEventProvider
}

// PlayResult carries the outcome of a completed play session, needed by the
// caller to record a score.
type PlayResult struct {
	Duration     float64
	LowestSpeed  float64
	Kills        int
	FreedMem     int64
	KillFailures []KillFailure
	Duds         []KillDud
}

// processKiller verifies and terminates a target's backing OS process, reporting
// whether the caller should reap the target because its process was already gone.
type processKiller interface {
	Kill(pid int, name string, protected bool) (shouldReap bool, err error)
}

// killSignal reports the outcome of a verified kill attempt: a real kill, a
// reap when the target's process had already exited, or a failure that
// leaves the target alive.
type killSignal struct {
	target     *game.Target
	shouldReap bool
	err        error // non-nil when the kill failed and the target must stay alive
}

// NewService constructs a Service with all required outbound ports injected.
func NewService(
	killer processKiller,
	renderer outbound.Renderer,
	events outbound.InputEventProvider,
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
func (s *Service) Play(cfg config.Config, processes []process.Info, highScore int) (PlayResult, error) {
	if err := process.ValidateProcesses(processes); err != nil {
		return PlayResult{}, apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityFatal, "", err)
	}

	session := game.NewSession(processes, game.Config{Confirm: cfg.ConfirmMode, Speed: cfg.Speed, TimeLimit: cfg.TimeLimit})

	tracker := newKillTracker(highScore)
	endTime, err := s.runLoop(session, tracker)
	if err != nil {
		return PlayResult{}, apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", err)
	}

	return PlayResult{
		Duration:     endTime.Sub(session.StartTime()).Seconds(),
		LowestSpeed:  session.Throttle().LowestSpeed(),
		Kills:        tracker.score.kills,
		FreedMem:     tracker.score.freedMem,
		KillFailures: tracker.failure.failures,
		Duds:         tracker.duds,
	}, nil
}

func (s *Service) runLoop(session *game.Session, tracker *killTracker) (time.Time, error) {
	if err := s.renderer.Init(); err != nil {
		return time.Time{}, fmt.Errorf("renderer initialization failed: %w", err)
	}
	defer s.renderer.Cleanup()

	session.Start(toBounds(s.renderer.WindowSize(), s.renderer.ChromeSize()))

	termSignal, stopWatching := s.registerTermSignalWatcher()
	defer stopWatching()

	// Closing done unblocks any killOrReap goroutine waiting to send on
	// killSignals after frameLoop exits, preventing a goroutine leak.
	done := make(chan struct{})
	defer close(done)

	dispatcher := input.NewDispatcher(game.NewInput(session))

	if err := s.frameLoop(session, tracker, dispatcher, termSignal, done); err != nil {
		return time.Time{}, err
	}

	// Capture end time before deferred cleanup runs.
	return time.Now(), nil
}

// registerTermSignalWatcher watches for an interrupt, termination, or suspend
// signal. termSignal is closed when that happens, letting callers react
// immediately instead of waiting out a polling interval. The returned stop watching
// func deregisters the watcher and must be called once the caller is done
// with game session.
func (s *Service) registerTermSignalWatcher() (termSignal <-chan struct{}, stopWatching func()) {
	ctx, stopWatching := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	return ctx.Done(), stopWatching
}

// frameLoop drives the game at frameDuration cadence until session stops running.
// It wakes immediately when session stops mid-frame or termSignal fires, instead of
// waiting out the remainder of the current tick.
func (s *Service) frameLoop(session *game.Session, tracker *killTracker, dispatcher *input.Dispatcher, termSignal <-chan struct{}, done <-chan struct{}) error {
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	killSignals := make(chan killSignal, 10)

	for session.IsRunning() {
		s.applyKillSignals(tracker, killSignals)
		if err := s.drainEventQueue(dispatcher, killSignals, done); err != nil {
			session.Stop()
			return err
		}
		window := s.renderer.WindowSize()
		session.Update(movement.WindowSize{Width: window.Width, Height: window.Height})
		s.renderer.Render(toFrameState(session, tracker))

		if !session.IsRunning() {
			break
		}

		select {
		case <-ticker.C:
		case <-termSignal:
			session.Stop()
		}
	}
	return nil
}

func (s *Service) applyKillSignals(tracker *killTracker, killSignals <-chan killSignal) {
	for {
		select {
		case sig := <-killSignals:
			sig.target.CeaseFire()
			switch {
			case sig.err != nil:
				tracker.recordFailure(sig.target, sig.err)
			case sig.shouldReap && sig.target.Reap():
				tracker.recordDud(sig.target)
			case sig.target.Kill():
				tracker.recordKill(sig.target.Rss)
			}
		default:
			return
		}
	}
}

// drainEventQueue dispatches every input event currently buffered, without
// blocking if none are ready. It returns an error if the event channel has
// closed, e.g. because the input adapter died unexpectedly — otherwise a
// closed channel is always ready to receive, and the loop below would spin
// forever redispatching its zero value instead of returning.
func (s *Service) drainEventQueue(dispatcher *input.Dispatcher, killSignals chan<- killSignal, done <-chan struct{}) error {
	events := s.events.Events()
	for {
		select {
		case inputEvent, ok := <-events:
			if !ok {
				return errors.New("input event channel closed")
			}
			if target := dispatcher.Dispatch(inputEvent); target != nil {
				target.FireShot()
				go s.killOrReap(target, killSignals, done)
			}
		default:
			return nil
		}
	}
}

// killOrReap verifies and kills target, then reports the outcome on
// killSignals: a real kill, a reap when the target's process had already
// exited, or a failure that leaves the target alive. The failure is not
// printed here — the renderer owns the terminal for the duration of the
// session, so the caller reports it only once the session has ended.
// killOrReap must be invoked via a goroutine: Kill may shell out to verify
// the target's backing process, and running it inline would stall the frame loop.
func (s *Service) killOrReap(target *game.Target, killSignals chan<- killSignal, done <-chan struct{}) {
	shouldReap, err := s.killer.Kill(target.PID, target.Name, target.Info.IsProtected())

	sig := killSignal{target: target}
	switch {
	case shouldReap:
		sig.shouldReap = true
	case err != nil:
		sig.err = err
	}

	select {
	case killSignals <- sig:
	case <-done:
	}
}
