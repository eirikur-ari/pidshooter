package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestNew(t *testing.T) {
	processes := []process.Info{
		{Pid: 1, Name: "a", Rss: 100},
		{Pid: 2, Name: "b", Rss: 200},
	}

	g := New(processes, Config{Confirm: true, Speed: 3.5, TimeLimit: 60})

	assert.True(t, g.cfg.Confirm)
	assert.Equal(t, 3.5, g.throttle.Speed())
	assert.Equal(t, 60, g.cfg.TimeLimit)
	assert.Equal(t, Pending, g.State())
	assert.Equal(t, 0, g.kills)
	assert.Equal(t, int64(0), g.freedMem)
}

func TestStart_TransitionsToRunning(t *testing.T) {
	processes := []process.Info{{Pid: 1, Name: "a", Rss: 100}}
	g := New(processes, Config{})

	g.Start(80, 24)

	assert.True(t, g.IsRunning())
	assert.Len(t, g.targets, 1)
}

func TestStop_TransitionsToStopped(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)

	g.Stop()

	assert.Equal(t, Stopped, g.State())
}

func TestStart_PanicsWhenRunning(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)

	assert.Panics(t, func() { g.Start(80, 24) })
}

func TestStart_PanicsWhenStopped(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)
	g.Stop()

	assert.Panics(t, func() { g.Start(80, 24) })
}

func TestGame_Speed(t *testing.T) {
	g := New(nil, Config{Speed: 2.5})
	assert.Equal(t, 2.5, g.Speed())
}

func TestGame_TimeLimit(t *testing.T) {
	g := New(nil, Config{TimeLimit: 30})
	assert.Equal(t, 30, g.TimeLimit())
}

func TestGame_Targets_EmptyBeforeStart(t *testing.T) {
	g := New([]process.Info{{Pid: 1, Name: "a"}}, Config{})
	assert.Empty(t, g.Targets())
}

func TestGame_Targets_PopulatedAfterStart(t *testing.T) {
	g := New([]process.Info{{Pid: 1, Name: "a"}}, Config{})
	g.Start(80, 24)
	assert.Len(t, g.Targets(), 1)
}

func TestGame_ConfirmTarget_NilWhenNoPending(t *testing.T) {
	g := New(nil, Config{})
	assert.Nil(t, g.ConfirmTarget())
}

func TestGame_ConfirmTarget_ReturnsPendingTarget(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 42, Name: "suspect"}}
	g := &Game{confirm: Confirmation{target: tgt, confirm: true}}
	assert.Equal(t, tgt, g.ConfirmTarget())
}

func TestGame_Confirm_PendingFalseInitially(t *testing.T) {
	g := New(nil, Config{})
	assert.False(t, g.Confirm().Pending())
}

func TestGame_Throttle_MutationAffectsSpeed(t *testing.T) {
	g := New(nil, Config{Speed: 2.0})
	g.Throttle().Increase()
	assert.Equal(t, 2.5, g.Speed())
}

func TestGame_Snapshot_ExcludesDeadTargets(t *testing.T) {
	g := New([]process.Info{{Pid: 1, Name: "a", Rss: 0}}, Config{Speed: 1.0})
	g.Start(80, 24)
	g.Kill(g.targets[0])
	for i := 0; i < KillAnimationDuration; i++ {
		g.Update(80, 24)
	}

	snap := g.Snapshot()

	assert.Empty(t, snap.Targets)
	assert.Equal(t, 0, snap.Alive)
}

func TestGame_Snapshot_CountsAlive(t *testing.T) {
	processes := []process.Info{
		{Pid: 1, Name: "a", Rss: 0},
		{Pid: 2, Name: "b", Rss: 0},
	}
	g := New(processes, Config{Speed: 1.0})
	g.Start(80, 24)
	g.Kill(g.targets[0])

	snap := g.Snapshot()

	require.Len(t, snap.Targets, 2) // killing + alive both visible
	assert.Equal(t, 1, snap.Alive)
}
