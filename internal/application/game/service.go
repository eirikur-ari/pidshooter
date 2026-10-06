package game

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// frameDuration is the time between frames.
const frameDuration = time.Second / 20

// defaultKillGracePeriod is how long a new Service waits for kills still in
// progress once the session has ended.
const defaultKillGracePeriod = 5 * time.Second

// Service plays game sessions.
type Service struct {
	killer   processKiller
	renderer outbound.Renderer
	events   outbound.InputEventProvider
	// killGracePeriod is how long to wait, once the session has ended, for
	// kills still in progress to finish before giving up on them.
	killGracePeriod time.Duration
}

// PlayRequest holds the parameters of a play session.
type PlayRequest struct {
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}

// PlayResult is the outcome of a completed play session.
type PlayResult struct {
	Duration     float64
	LowestSpeed  float64
	Kills        int
	FreedMem     int64
	KillFailures []KillFailure
	Duds         []KillDud
}

// processKiller kills the process behind a target. shouldReap is true when
// that process was already gone, so the target should be reaped instead.
type processKiller interface {
	Kill(pid int, name string, protected bool) (shouldReap bool, err error)
}

// killSignal is the outcome of one kill attempt on a target.
type killSignal struct {
	target     *game.Target
	shouldReap bool
	err        error // non-nil when the kill failed and the target must stay alive
}

// NewService returns a Service that kills through killer, draws with
// renderer, and reads input from events.
func NewService(
	killer processKiller,
	renderer outbound.Renderer,
	events outbound.InputEventProvider,
) *Service {
	return &Service{
		killer:          killer,
		renderer:        renderer,
		events:          events,
		killGracePeriod: defaultKillGracePeriod,
	}
}

// Play runs a game session over processes. highScore is the best score so
// far, and is raised during the session as kills exceed it.
func (s *Service) Play(req PlayRequest, processes []process.Info, highScore int) (PlayResult, error) {
	if err := validateGameConfig(req); err != nil {
		return PlayResult{}, apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", err)
	}
	if err := process.ValidateProcesses(processes); err != nil {
		return PlayResult{}, apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityFatal, "", err)
	}

	session := game.NewSession(processes, toGameConfig(req))

	tracker := newKillTracker(highScore)
	endTime, err := s.runLoop(session, tracker)
	if err != nil {
		return PlayResult{}, apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", err)
	}

	return PlayResult{
		Duration:     endTime.Sub(session.StartTime()).Seconds(),
		LowestSpeed:  session.Throttle().LowestSpeed(),
		Kills:        tracker.kills(),
		FreedMem:     tracker.freedMem(),
		KillFailures: tracker.failures(),
		Duds:         tracker.duds,
	}, nil
}

// validateGameConfig returns an error if the request's speed or time limit is invalid.
func validateGameConfig(req PlayRequest) error {
	if err := movement.ValidateSpeed(req.Speed); err != nil {
		return err
	}
	if err := game.ValidateTimeLimit(req.TimeLimit); err != nil {
		return err
	}
	return nil
}

func (s *Service) runLoop(session *game.Session, tracker *killTracker) (time.Time, error) {
	if err := s.renderer.Init(); err != nil {
		return time.Time{}, fmt.Errorf("renderer initialization failed: %w", err)
	}
	defer s.renderer.Cleanup()

	session.Start(toBounds(s.renderer.WindowSize(), s.renderer.ChromeSize()))

	termSignal, stopWatching := s.registerTermSignalWatcher()
	defer stopWatching()

	dispatcher := input.NewDispatcher(game.NewInput(session))

	return s.frameLoop(session, tracker, dispatcher, termSignal)
}

// registerTermSignalWatcher returns a channel that closes when an interrupt,
// termination, or suspend signal arrives, and a func that stops watching.
func (s *Service) registerTermSignalWatcher() (termSignal <-chan struct{}, stopWatching func()) {
	ctx, stopWatching := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	return ctx.Done(), stopWatching
}

// frameLoop runs frames until the session ends or a termination signal
// arrives, then waits for in-flight kills. It returns the time the loop
// ended, before that wait.
func (s *Service) frameLoop(session *game.Session, tracker *killTracker, dispatcher *input.Dispatcher, termSignal <-chan struct{}) (time.Time, error) {
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	killSignals := make(chan killSignal, 10) // channel buffer capacity = 10 kills in flight
	var waitGroup sync.WaitGroup

	// Closing done unblocks any killOrReap goroutine waiting to send on
	// killSignals after frameLoop exits, preventing a goroutine leak.
	done := make(chan struct{})
	defer close(done)

	for session.IsRunning() {
		s.applyKillSignals(tracker, killSignals)
		if err := s.drainEventQueue(dispatcher, killSignals, done, &waitGroup); err != nil {
			session.Stop()
			return time.Time{}, err
		}
		session.Update(toWindowSize(s.renderer.WindowSize()))
		s.renderer.Render(toFrameViewState(session, tracker))

		if !session.IsRunning() {
			break
		}

		select {
		case <-ticker.C:
		case <-termSignal:
			session.Stop()
		}
	}

	endTime := time.Now()
	s.awaitOutstandingKills(&waitGroup, tracker, killSignals)
	return endTime, nil
}

// awaitOutstandingKills applies the outcomes of kills still in progress,
// waiting at most the grace period. Later outcomes are dropped.
func (s *Service) awaitOutstandingKills(waitGroup *sync.WaitGroup, tracker *killTracker, killSignals chan killSignal) {
	awaited := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(awaited)
	}()

	timer := time.NewTimer(s.killGracePeriod)
	defer timer.Stop()

	for {
		select {
		case sig := <-killSignals:
			s.applyKillSignal(tracker, sig)
		case <-awaited:
			s.applyKillSignals(tracker, killSignals)
			return
		case <-timer.C:
			s.applyKillSignals(tracker, killSignals)
			return
		}
	}
}

// applyKillSignals applies every signal currently buffered, without blocking.
func (s *Service) applyKillSignals(tracker *killTracker, killSignals <-chan killSignal) {
	for {
		select {
		case sig := <-killSignals:
			s.applyKillSignal(tracker, sig)
		default:
			return
		}
	}
}

// applyKillSignal applies a kill outcome to its target and the tracker.
func (s *Service) applyKillSignal(tracker *killTracker, sig killSignal) {
	sig.target.CeaseFire()
	switch {
	case sig.err != nil:
		tracker.recordFailure(sig.target, sig.err)
	case sig.shouldReap:
		if sig.target.Reap() {
			tracker.recordDud(sig.target)
		}
	default:
		if sig.target.Kill() {
			tracker.recordKill(sig.target.Info.Rss)
		}
	}
}

// drainEventQueue dispatches every input event currently buffered, without
// blocking, and starts a kill for each target hit. It returns an error if
// the event channel is closed.
func (s *Service) drainEventQueue(dispatcher *input.Dispatcher, killSignals chan<- killSignal, done <-chan struct{}, waitGroup *sync.WaitGroup) error {
	events := s.events.Events()
	for {
		select {
		case inputEvent, ok := <-events:
			if !ok {
				return errors.New("input event channel closed")
			}
			if target := dispatcher.Dispatch(inputEvent); target != nil {
				target.FireShot()
				waitGroup.Add(1)
				go s.killOrReap(target, killSignals, done, waitGroup)
			}
		default:
			return nil
		}
	}
}

// killOrReap kills target and sends the outcome on killSignals, unless done
// closes first. Run it in a goroutine after waitGroup.Add(1); it calls
// waitGroup.Done on return.
func (s *Service) killOrReap(target *game.Target, killSignals chan<- killSignal, done <-chan struct{}, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	shouldReap, err := s.killer.Kill(target.Info.PID, target.Info.Name, target.Info.IsProtected())

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
