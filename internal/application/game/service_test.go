package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

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

// --- registerTermSignalWatcher ---

func TestRegisterTermSignalWatcherTermSignalClosesOnStop(t *testing.T) {
	session := game.NewSession(nil, game.Config{Speed: 1.0})
	session.Start(80, 24)
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputSource())

	termSignal, stop := svc.registerTermSignalWatcher(session)
	stop()

	select {
	case <-termSignal:
	case <-time.After(time.Second):
		t.Fatal("termSignal was not closed after stop()")
	}
}
