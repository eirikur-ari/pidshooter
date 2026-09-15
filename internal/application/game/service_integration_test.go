//go:build integration

package game

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func TestIntegrationServiceFrameLoopAppliesAsyncKillToResult(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	session := game.NewSession([]process.Info{info}, game.Config{Speed: 1.0})
	session.Start(80, 24)

	target := session.Targets()[0]
	x, y := int(target.Position.X), int(target.Position.Y)

	events := fake.NewInputSource()
	events.Ch <- outbound.ClickEvent{X: x, Y: y}

	killer := fake.ProcessKiller(func(int, string, bool) (bool, error) {
		return false, nil
	})
	svc := NewService(killer, &fake.Renderer{}, events)
	dispatcher := event.NewDispatcher(game.NewInput(session))

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

func TestIntegrationServiceDrainEventQueueIgnoresDuplicateClicksOnSameTarget(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	session := game.NewSession([]process.Info{info}, game.Config{Speed: 1.0})
	session.Start(80, 24)

	target := session.Targets()[0]
	x, y := int(target.Position.X), int(target.Position.Y)

	events := fake.NewInputSource()
	events.Ch <- outbound.ClickEvent{X: x, Y: y}
	events.Ch <- outbound.ClickEvent{X: x, Y: y}

	var calls atomic.Int32
	killer := fake.ProcessKiller(func(int, string, bool) (bool, error) {
		calls.Add(1)
		return false, nil
	})
	svc := NewService(killer, &fake.Renderer{}, events)
	dispatcher := event.NewDispatcher(game.NewInput(session))

	killSignals := make(chan killSignal, 10)
	done := make(chan struct{})
	defer close(done)

	svc.drainEventQueue(dispatcher, killSignals, done)

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
