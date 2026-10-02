//go:build integration

package game

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func TestIntegrationServiceFrameLoopAppliesAsyncKillToResult(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	session := game.NewSession([]process.Info{info}, game.Config{Speed: 1.0})
	session.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))

	target := session.Targets()[0]
	x, y := target.Motion.Position.Rounded()

	events := fake.NewInputEventProvider()
	events.Ch <- input.ClickEvent{X: x, Y: y}

	killer := fake.ProcessKiller(func(int, string, bool) (bool, error) {
		return false, nil
	})
	svc := NewService(killer, &fake.Renderer{}, events)
	dispatcher := input.NewDispatcher(game.NewInput(session))

	done := make(chan struct{})
	defer close(done)

	tracker := newKillTracker(0)
	finished := make(chan struct{})
	go func() {
		svc.frameLoop(session, tracker, dispatcher, nil, done)
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(6 * time.Second):
		t.Fatal("frameLoop did not finish the kill animation within 6s")
	}

	assert.Equal(t, 1, tracker.score.kills)
	assert.Equal(t, info.Rss, tracker.score.freedMem)
	assert.Empty(t, tracker.failure.failures)
}

func TestIntegrationServiceFrameLoopWaitsForKillInFlightWhenSessionStops(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	session := game.NewSession([]process.Info{info}, game.Config{Speed: 1.0})
	session.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))

	target := session.Targets()[0]
	x, y := target.Motion.Position.Rounded()

	events := fake.NewInputEventProvider()
	events.Ch <- input.ClickEvent{X: x, Y: y}
	events.Ch <- input.QuitEvent{}

	const killDelay = 300 * time.Millisecond
	killer := fake.ProcessKiller(func(int, string, bool) (bool, error) {
		time.Sleep(killDelay)
		return false, nil
	})
	svc := NewService(killer, &fake.Renderer{}, events)
	dispatcher := input.NewDispatcher(game.NewInput(session))

	done := make(chan struct{})
	defer close(done)

	tracker := newKillTracker(0)
	start := time.Now()
	svc.frameLoop(session, tracker, dispatcher, nil, done)

	assert.GreaterOrEqual(t, time.Since(start), killDelay, "frameLoop must wait for an in-flight kill before returning")
	assert.Equal(t, 1, tracker.score.kills, "a kill that lands after the session stops must still be counted")
}

func TestIntegrationAwaitOutstandingKillsReturnsOnceGracePeriodElapses(t *testing.T) {
	svc := &Service{killGracePeriod: 20 * time.Millisecond}
	var waitGroup sync.WaitGroup
	waitGroup.Add(1) // never Done() — simulates a permanently wedged killOrReap goroutine

	killSignals := make(chan killSignal, 1)
	tracker := newKillTracker(0)

	start := time.Now()
	svc.awaitOutstandingKills(&waitGroup, tracker, killSignals)
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 20*time.Millisecond)
	assert.Less(t, elapsed, time.Second, "awaitOutstandingKills must give up once the grace period elapses, not hang indefinitely")
}

func TestIntegrationServiceDrainEventQueueIgnoresDuplicateClicksOnSameTarget(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	session := game.NewSession([]process.Info{info}, game.Config{Speed: 1.0})
	session.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))

	target := session.Targets()[0]
	x, y := target.Motion.Position.Rounded()

	events := fake.NewInputEventProvider()
	events.Ch <- input.ClickEvent{X: x, Y: y}
	events.Ch <- input.ClickEvent{X: x, Y: y}

	var calls atomic.Int32
	killer := fake.ProcessKiller(func(int, string, bool) (bool, error) {
		calls.Add(1)
		return false, nil
	})
	svc := NewService(killer, &fake.Renderer{}, events)
	dispatcher := input.NewDispatcher(game.NewInput(session))

	killSignals := make(chan killSignal, 10)
	done := make(chan struct{})
	defer close(done)

	svc.drainEventQueue(dispatcher, killSignals, done, &sync.WaitGroup{})

	select {
	case sig := <-killSignals:
		assert.Equal(t, target, sig.target)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the kill signal from the first click")
	}

	select {
	case <-killSignals:
		t.Fatal("a second click on the same target while its first kill attempt is in flight must not dispatch a second kill")
	case <-time.After(200 * time.Millisecond):
	}

	assert.Equal(t, int32(1), calls.Load())
}
