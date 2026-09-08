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
	tgt := game.NewTarget(process.NewInfo(42, "dummy", 0, 0), movement.NewBounds(80, 24))
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
}

// --- toTargetViewStates ---

func TestToTargetViewStatesReturnsASliceOfMappedFields(t *testing.T) {
	targets := []*game.Target{
		{Info: process.NewInfo(1, "a", 0, 0), State: game.Alive},
		{Info: process.NewInfo(2, "b", 0, 0), State: game.Killing},
	}

	views := toTargetViewStates(targets)

	require.Len(t, views, 2)
	assert.False(t, views[0].Killing)
	assert.True(t, views[1].Killing)
}

func TestToTargetViewStatesReturnsEmptySliceWhenInputIsNil(t *testing.T) {
	assert.Empty(t, toTargetViewStates(nil))
}

// --- toHUDState ---

func TestToHUDStateReturnsMappedFields(t *testing.T) {
	tracker := newKillTracker(5)
	tracker.recordKill(2048)
	tracker.recordKill(2048)

	hud := toHUDState(tracker)

	assert.Equal(t, int64(4096), hud.FreedMem)
	assert.Equal(t, 2, hud.Kills)
	assert.Equal(t, 5, hud.HighScore)
}

// --- toStatusState ---

func TestToStatusStateReturnsMappedFieldsWithoutConfirmViewState(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(1, "a", 0, 0)}, game.Config{Speed: 2.0, TimeLimit: 30})
	session.Start(80, 24)

	status := toStatusState(session, 3)

	assert.Equal(t, 3, status.Alive)
	assert.Equal(t, 2.0, status.Speed)
	assert.Equal(t, 30, status.TimeLimit)
	assert.Nil(t, status.Confirming)
}

func TestToStatusStateIncludesConfirmViewState(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(42, "suspect", 0, 0)}, game.Config{Confirm: true, Speed: 1.0})
	session.Start(80, 24)
	session.RequestConfirm(session.Targets()[0])

	status := toStatusState(session, 1)

	require.NotNil(t, status.Confirming)
	assert.Equal(t, 42, status.Confirming.PID)
}

// --- toFrameState ---

func TestToFrameStateReturnsMappedFields(t *testing.T) {
	session := game.NewSession([]process.Info{process.NewInfo(1, "a", 0, 0)}, game.Config{Speed: 1.0})
	session.Start(80, 24)

	f := toFrameState(session, newKillTracker(0))

	require.Len(t, f.Targets, 1)
	assert.False(t, f.Targets[0].Killing)
}
