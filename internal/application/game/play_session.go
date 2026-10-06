package game

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

// frameDuration is the time between frames.
const frameDuration = time.Second / 20

// playSession is a single play of the game, from setup to result.
type playSession struct {
	renderer   outbound.Renderer
	events     outbound.InputEventProvider
	session    *game.Session
	dispatcher *input.Dispatcher
	tracker    *killTracker
	kills      *inFlightKills
}

// newPlaySession returns a playSession that plays session and records its
// progress in tracker, or an error if either is nil. Kills go through killer,
// and killGracePeriod is how long to wait for kills still in progress once
// the session has ended.
func newPlaySession(
	renderer outbound.Renderer,
	events outbound.InputEventProvider,
	killer processKiller,
	killGracePeriod time.Duration,
	session *game.Session,
	tracker *killTracker,
) (*playSession, error) {
	if session == nil {
		return nil, errors.New("game session is required")
	}
	if tracker == nil {
		return nil, errors.New("kill tracker is required")
	}

	return &playSession{
		renderer:   renderer,
		events:     events,
		session:    session,
		dispatcher: input.NewDispatcher(game.NewInput(session)),
		tracker:    tracker,
		kills:      newInFlightKills(killer, killGracePeriod, tracker),
	}, nil
}

// run plays the session and returns the time its frame loop ended.
func (p *playSession) run() (time.Time, error) {
	if err := p.renderer.Init(); err != nil {
		return time.Time{}, fmt.Errorf("renderer initialization failed: %w", err)
	}
	defer p.renderer.Cleanup()

	p.session.Start(toBounds(p.renderer.WindowSize(), p.renderer.ChromeSize()))

	termSignal, stopWatching := registerTermSignalWatcher()
	defer stopWatching()

	defer p.kills.close()

	return p.frameLoop(termSignal)
}

// result returns the outcome of the session, given the time its frame loop ended.
func (p *playSession) result(endTime time.Time) PlayResult {
	return PlayResult{
		Duration:     endTime.Sub(p.session.StartTime()).Seconds(),
		LowestSpeed:  p.session.Throttle().LowestSpeed(),
		Kills:        p.tracker.kills(),
		FreedMem:     p.tracker.freedMem(),
		KillFailures: p.tracker.failures(),
		Duds:         p.tracker.duds,
	}
}

// frameLoop runs frames until the session ends or a termination signal
// arrives, then waits for in-flight kills. It returns the time the loop
// ended, before that wait.
func (p *playSession) frameLoop(termSignal <-chan struct{}) (time.Time, error) {
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	for p.session.IsRunning() {
		p.kills.applyFinished()
		if err := p.drainEventQueue(); err != nil {
			p.session.Stop()
			return time.Time{}, err
		}
		p.session.Update(toWindowSize(p.renderer.WindowSize()))
		p.renderer.Render(toFrameViewState(p.session, p.tracker))

		if !p.session.IsRunning() {
			break
		}

		select {
		case <-ticker.C:
		case <-termSignal:
			p.session.Stop()
		}
	}

	endTime := time.Now()
	p.kills.awaitRemaining()
	return endTime, nil
}

// drainEventQueue dispatches every input event currently buffered, without
// blocking, and starts a kill for each target hit. It returns an error if
// the event channel is closed.
func (p *playSession) drainEventQueue() error {
	events := p.events.Events()
	for {
		select {
		case inputEvent, ok := <-events:
			if !ok {
				return errors.New("input event channel closed")
			}
			if target := p.dispatcher.Dispatch(inputEvent); target != nil {
				p.kills.start(target)
			}
		default:
			return nil
		}
	}
}

// registerTermSignalWatcher returns a channel that closes when an interrupt,
// termination, or suspend signal arrives, and a func that stops watching.
func registerTermSignalWatcher() (termSignal <-chan struct{}, stopWatching func()) {
	ctx, stopWatching := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	return ctx.Done(), stopWatching
}
