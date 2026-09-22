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

func TestToConfirmViewStateReturnNilWhenInputIsNil(t *testing.T) {
	assert.Nil(t, toConfirmViewState(nil))
}

func TestToConfirmViewStateReturnsMappedFields(t *testing.T) {
	tgt := game.NewTarget(process.NewInfo(42, "dummy", 0, 0), movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	vs := toConfirmViewState(tgt)
	require.NotNil(t, vs)
	assert.Equal(t, 42, vs.PID)
	assert.Equal(t, "dummy", vs.Name)
}

// --- toTargetViewState ---

func TestToTargetViewStateReturnsMappedFields(t *testing.T) {
	tgt := &game.Target{
		Info:   process.NewInfo(42, "dummy", 0, 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10.6, Y: 5.4}},
		State:  game.Alive,
	}

	view := toTargetViewState(tgt)

	assert.Equal(t, 11, view.X)
	assert.Equal(t, 5, view.Y)
	assert.Equal(t, "[42 dummy]", view.Tag)
	assert.False(t, view.Killing)
	assert.Equal(t, 0.0, view.AnimationProgress)
}

func TestToTargetViewStateIncludesAnimationProgressWhenKilling(t *testing.T) {
	tgt := &game.Target{
		Info:          process.NewInfo(42, "dummy", 0, 0),
		State:         game.Killing,
		AnimationTick: game.AnimationDuration / 2,
	}

	view := toTargetViewState(tgt)

	assert.True(t, view.Killing)
	assert.Equal(t, 0.5, view.AnimationProgress)
}

func TestToTargetViewStatesReturnsASliceOfMappedFields(t *testing.T) {
	targets := []*game.Target{
		{Info: process.NewInfo(1, "a", 0, 0), State: game.Alive},
		{Info: process.NewInfo(2, "b", 0, 0), State: game.Killing},
		{Info: process.NewInfo(3, "c", 0, 0), State: game.Fleeing},
	}

	views := toTargetViewStates(targets)

	require.Len(t, views, 3)
	assert.False(t, views[0].Killing)
	assert.False(t, views[0].Fleeing)
	assert.True(t, views[1].Killing)
	assert.False(t, views[1].Fleeing)
	assert.False(t, views[2].Killing)
	assert.True(t, views[2].Fleeing)
}

func TestToTargetViewStatesReturnsEmptySliceWhenInputIsNil(t *testing.T) {
	assert.Empty(t, toTargetViewStates(nil))
}

// --- toTargetViewStates ---

// --- toHUDViewState ---

func TestToHUDViewStateReturnsMappedFields(t *testing.T) {
	tracker := newKillTracker(5)
	tracker.recordKill(2048)
	tracker.recordKill(2048)

	hud := toHUDViewState(tracker)

	assert.Equal(t, int64(4096), hud.FreedMem)
	assert.Equal(t, 2, hud.Kills)
	assert.Equal(t, 5, hud.HighScore)
}

// --- toStatusViewState ---

func TestToStatusViewStateReturnsMappedFieldsWithoutConfirmViewState(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(1, "a", 0, 0)}, game.Config{Speed: 2.0, TimeLimit: 30})
	session.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))

	status := toStatusViewState(session, 3)

	assert.Equal(t, 3, status.Alive)
	assert.Equal(t, 2.0, status.Speed)
	require.NotNil(t, status.TimeLeft)
	assert.Equal(t, 30, *status.TimeLeft)
	assert.Nil(t, status.Confirming)
}

func TestToStatusViewStateTimeLeftNilWhenUntimed(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(1, "a", 0, 0)}, game.Config{Speed: 2.0})
	session.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))

	status := toStatusViewState(session, 1)

	assert.Nil(t, status.TimeLeft)
}

func TestToStatusViewStateIncludesConfirmViewState(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(42, "suspect", 0, 0)}, game.Config{Confirm: true, Speed: 1.0})
	session.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	session.RequestConfirm(session.Targets()[0])

	status := toStatusViewState(session, 1)

	require.NotNil(t, status.Confirming)
	assert.Equal(t, 42, status.Confirming.PID)
}

// --- toFrameViewState ---

func TestToFrameViewStateReturnsMappedFields(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(1, "a", 0, 0)}, game.Config{Speed: 1.0})
	session.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))

	f := toFrameViewState(session, newKillTracker(0))

	require.Len(t, f.Targets, 1)
	assert.False(t, f.Targets[0].Killing)
}
