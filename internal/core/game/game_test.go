package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestNew(t *testing.T) {
	processes := []process.Info{
		process.NewInfo(1, "a", 100),
		process.NewInfo(2, "b", 200),
	}

	g := New(processes, Config{Confirm: true, Speed: 3.5, TimeLimit: 60})

	assert.True(t, g.cfg.Confirm)
	assert.Equal(t, 3.5, g.throttle.Speed())
	assert.Equal(t, 60, g.cfg.TimeLimit)
	assert.Equal(t, pending, g.currentState())
}

func TestStartTransitionsToRunning(t *testing.T) {
	processes := []process.Info{process.NewInfo(1, "a", 100)}
	g := New(processes, Config{})

	g.Start(80, 24)

	assert.True(t, g.IsRunning())
	assert.Len(t, g.targets, 1)
}

func TestStopTransitionsToStopped(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)

	g.Stop()

	assert.Equal(t, stopped, g.currentState())
}

func TestStartPanicsWhenRunning(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)

	assert.Panics(t, func() { g.Start(80, 24) })
}

func TestStartPanicsWhenStopped(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)
	g.Stop()

	assert.Panics(t, func() { g.Start(80, 24) })
}

func TestGameTimeLimit(t *testing.T) {
	g := New(nil, Config{TimeLimit: 30})
	assert.Equal(t, 30, g.TimeLimit())
}

func TestGameTargetsEmptyBeforeStart(t *testing.T) {
	g := New([]process.Info{process.NewInfo(1, "a", 0)}, Config{})
	assert.Empty(t, g.Targets())
}

func TestGameTargetsPopulatedAfterStart(t *testing.T) {
	g := New([]process.Info{process.NewInfo(1, "a", 0)}, Config{})
	g.Start(80, 24)
	assert.Len(t, g.Targets(), 1)
}

func TestGameConfirmTargetNilWhenNoPending(t *testing.T) {
	g := New(nil, Config{})
	assert.Nil(t, g.ConfirmTarget())
}

func TestGameConfirmTargetReturnsPendingTarget(t *testing.T) {
	tgt := &Target{Info: process.NewInfo(42, "suspect", 0)}
	g := &Game{confirm: Confirmation{target: tgt, confirm: true}}
	assert.Equal(t, tgt, g.ConfirmTarget())
}

func TestGameConfirmPendingFalseInitially(t *testing.T) {
	g := New(nil, Config{})
	assert.False(t, g.Confirm().Pending())
}

func TestGameThrottleMutationAffectsSpeed(t *testing.T) {
	g := New(nil, Config{Speed: 2.0})
	g.Throttle().Increase()
	assert.Equal(t, 2.5, g.Throttle().Speed())
}

func TestGameFrameExcludesDeadTargets(t *testing.T) {
	g := New([]process.Info{process.NewInfo(1, "a", 0)}, Config{Speed: 1.0})
	g.Start(80, 24)
	g.targets[0].Kill()
	for range KillAnimationDuration {
		g.Update(80, 24)
	}

	frame := g.Frame()

	assert.Empty(t, frame.Targets)
	assert.Equal(t, 0, frame.Alive)
}

func TestGameFrameCountsAlive(t *testing.T) {
	processes := []process.Info{
		process.NewInfo(1, "a", 0),
		process.NewInfo(2, "b", 0),
	}
	g := New(processes, Config{Speed: 1.0})
	g.Start(80, 24)
	g.targets[0].Kill()

	frame := g.Frame()

	require.Len(t, frame.Targets, 2) // killing + alive both visible
	assert.Equal(t, 1, frame.Alive)
}
