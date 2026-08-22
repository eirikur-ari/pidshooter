package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// killerFunc adapts a plain function to the killer interface for tests.
type killerFunc func(target *game.Target) (killed, shouldReap bool, err error)

func (f killerFunc) Kill(target *game.Target) (killed, shouldReap bool, err error) {
	return f(target)
}

// --- applyKillSignals ---

func TestServiceApplyKillsCompletesPendingKill(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputSource())
	kills := make(chan killSignal, 1)

	target := game.NewTarget(info, movement.NewBounds(80, 24))
	kills <- killSignal{target: target}

	tracker := &scoreTracker{}
	svc.applyKillSignals(tracker, kills)

	assert.Equal(t, game.Killing, target.State)
	assert.Equal(t, 1, tracker.kills)
}

func TestServiceApplyKillsReapsAlreadyKilledTarget(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputSource())
	kills := make(chan killSignal, 1)

	target := game.NewTarget(info, movement.NewBounds(80, 24))
	kills <- killSignal{target: target, shouldReap: true}

	tracker := &scoreTracker{}
	svc.applyKillSignals(tracker, kills)

	assert.Equal(t, game.Dead, target.State)
	assert.Equal(t, 0, tracker.kills, "reaping an already-gone target should not award a kill")
}

func TestServiceApplyKillsEmptyChannelNoOps(t *testing.T) {
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputSource())
	kills := make(chan killSignal, 1)

	tracker := &scoreTracker{}
	svc.applyKillSignals(tracker, kills) // must not block

	assert.Equal(t, 0, tracker.kills)
}

// --- frameLoop (async kill end-to-end) ---

// TestServiceFrameLoopAppliesAsyncKillToResult drives a click through the real
// dispatch -> killOrReap goroutine -> applyKillSignals pipeline that Play
// relies on, verifying a confirmed kill is reflected in the tracked result.
func TestServiceFrameLoopAppliesAsyncKillToResult(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	session := game.NewSession([]process.Info{info}, game.Config{Speed: 1.0})
	session.Start(80, 24)

	target := session.Targets()[0]
	x, y := int(target.Position.X), int(target.Position.Y)

	events := fake.NewInputSource()
	events.Ch <- outbound.ClickEvent{X: x, Y: y}

	killer := killerFunc(func(*game.Target) (bool, bool, error) {
		return true, false, nil
	})
	svc := NewService(killer, &fake.Renderer{}, events)
	dispatcher := event.NewDispatcher(game.NewInput(session))

	done := make(chan struct{})
	defer close(done)

	tracker := &scoreTracker{}
	svc.frameLoop(session, tracker, dispatcher, nil, done)

	assert.Equal(t, 1, tracker.kills)
	assert.Equal(t, info.Rss, tracker.freedMem)
}

// --- registerTermSignalWatcher ---

func TestRegisterTermSignalWatcherTermSignalClosesOnStop(t *testing.T) {
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputSource())

	termSignal, stop := svc.registerTermSignalWatcher()
	stop()

	select {
	case <-termSignal:
	case <-time.After(time.Second):
		t.Fatal("termSignal was not closed after stop()")
	}
}
