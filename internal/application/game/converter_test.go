package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// --- toConfirmViewState ---

func TestToConfirmViewStateNilInput(t *testing.T) {
	assert.Nil(t, toConfirmViewState(nil))
}

func TestToConfirmViewStateMapsFields(t *testing.T) {
	tgt := game.NewTarget(process.NewInfo(42, "suspect", 0), movement.NewBounds(80, 24))
	vs := toConfirmViewState(tgt)
	require.NotNil(t, vs)
	assert.Equal(t, 42, vs.Pid)
	assert.Equal(t, "suspect", vs.Name)
}

// --- toTargetViewState ---

func TestToTargetViewStateRoundsPosition(t *testing.T) {
	tgt := &game.Target{
		Info:   process.NewInfo(1, "x", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10.6, Y: 5.4}},
		State:  game.Alive,
	}

	view := toTargetViewState(tgt)

	assert.Equal(t, 11, view.X)
	assert.Equal(t, 5, view.Y)
	assert.Equal(t, "[1 x]", view.Tag)
	assert.False(t, view.Killing)
}

func TestToTargetViewStateMarksKilling(t *testing.T) {
	tgt := &game.Target{Info: process.NewInfo(1, "x", 0), State: game.Killing}

	view := toTargetViewState(tgt)

	assert.True(t, view.Killing)
}

// --- toTargetViewStates ---

func TestToTargetViewStatesMapsAll(t *testing.T) {
	targets := []*game.Target{
		{Info: process.NewInfo(1, "a", 0), State: game.Alive},
		{Info: process.NewInfo(2, "b", 0), State: game.Killing},
	}

	views := toTargetViewStates(targets)

	require.Len(t, views, 2)
	assert.False(t, views[0].Killing)
	assert.True(t, views[1].Killing)
}

func TestToTargetViewStatesEmptyInput(t *testing.T) {
	assert.Empty(t, toTargetViewStates(nil))
}

// --- toHUDState ---

func TestToHUDStateMapsFields(t *testing.T) {
	tracker := &scoreTracker{freedMem: 4096, kills: 2, highScore: 5}

	hud := toHUDState(tracker)

	assert.Equal(t, int64(4096), hud.FreedMem)
	assert.Equal(t, 2, hud.Kills)
	assert.Equal(t, 5, hud.HighScore)
}

// --- toStatusState ---

func TestToStatusStateMapsFields(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(1, "a", 0)}, game.Config{Speed: 2.0, TimeLimit: 30})
	session.Start(80, 24)

	status := toStatusState(session, 3)

	assert.Equal(t, 3, status.Alive)
	assert.Equal(t, 2.0, status.Speed)
	assert.Equal(t, 30, status.TimeLimit)
	assert.Nil(t, status.Confirming)
}

func TestToStatusStateIncludesConfirming(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(42, "suspect", 0)}, game.Config{Confirm: true, Speed: 1.0})
	session.Start(80, 24)
	session.RequestConfirm(session.Targets()[0])

	status := toStatusState(session, 1)

	require.NotNil(t, status.Confirming)
	assert.Equal(t, 42, status.Confirming.Pid)
}

// --- toFrameState ---

func TestToFrameStateAliveTargetIncluded(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(1, "a", 0)}, game.Config{Speed: 1.0})
	session.Start(80, 24)

	f := toFrameState(session, &scoreTracker{})

	require.Len(t, f.Targets, 1)
	assert.False(t, f.Targets[0].Killing)
}

func TestToFrameStateKillingTargetMarked(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(1, "a", 0)}, game.Config{Speed: 1.0})
	session.Start(80, 24)
	session.Targets()[0].Kill()

	f := toFrameState(session, &scoreTracker{})

	require.Len(t, f.Targets, 1)
	assert.True(t, f.Targets[0].Killing)
}

func TestToFrameStateDeadTargetExcluded(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(1, "a", 0)}, game.Config{Speed: 1.0})
	session.Start(80, 24)
	session.Targets()[0].Kill()
	for i := 0; i < game.KillAnimationDuration; i++ {
		session.Update(80, 24)
	}

	f := toFrameState(session, &scoreTracker{})

	assert.Empty(t, f.Targets)
}

func TestToFrameStateHUDReflectsStats(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(1, "a", 4096)}, game.Config{Speed: 1.0})
	session.Start(80, 24)
	tracker := &scoreTracker{}
	tracker.recordKill(session.Targets()[0].Rss)

	f := toFrameState(session, tracker)

	assert.Equal(t, 1, f.HUD.Kills)
	assert.Equal(t, int64(4096), f.HUD.FreedMem)
}

func TestToFrameStateStatusBarAliveCount(t *testing.T) {
	processes := []process.Info{
		process.NewInfo(1, "a", 0),
		process.NewInfo(2, "b", 0),
	}
	session := game.NewSession(processes, game.Config{Speed: 1.0})
	session.Start(80, 24)
	session.Targets()[0].Kill()

	f := toFrameState(session, &scoreTracker{})

	assert.Equal(t, 1, f.StatusBar.Alive)
}

func TestToFrameStateNoConfirmPending(t *testing.T) {
	session := game.NewSession(nil, game.Config{Speed: 1.0})
	session.Start(80, 24)

	f := toFrameState(session, &scoreTracker{})

	assert.Nil(t, f.StatusBar.Confirming)
}
