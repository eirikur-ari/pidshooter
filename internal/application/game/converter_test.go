package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

// --- toProcessInfo ---

func TestToProcessInfoMapsFields(t *testing.T) {
	info := toProcessInfo(outbound.ProcessInfo{Pid: 42, Name: "suspect", Rss: 1024})

	assert.Equal(t, 42, info.Pid)
	assert.Equal(t, "suspect", info.Name)
	assert.Equal(t, int64(1024), info.Rss)
}

// --- toProcessInfos ---

func TestToProcessInfosMapsAll(t *testing.T) {
	infos := toProcessInfos([]outbound.ProcessInfo{
		{Pid: 1, Name: "a", Rss: 100},
		{Pid: 2, Name: "b", Rss: 200},
	})

	require.Len(t, infos, 2)
	assert.Equal(t, 1, infos[0].Pid)
	assert.Equal(t, "a", infos[0].Name)
	assert.Equal(t, int64(100), infos[0].Rss)
	assert.Equal(t, 2, infos[1].Pid)
	assert.Equal(t, "b", infos[1].Name)
	assert.Equal(t, int64(200), infos[1].Rss)
}

func TestToProcessInfosEmptyInput(t *testing.T) {
	infos := toProcessInfos(nil)

	assert.Empty(t, infos)
}

// --- toConfirmViewState ---

func TestToConfirmViewStateNilInput(t *testing.T) {
	assert.Nil(t, toConfirmViewState(nil))
}

func TestToConfirmViewStateMapsFields(t *testing.T) {
	tgt := game.NewTarget(process.NewInfo(42, "suspect", 0), movement.NewBounds(80, 24))
	vs := toConfirmViewState(tgt)
	require.NotNil(t, vs)
	assert.Equal(t, 42, vs.PID)
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
	tracker := &score.Tracker{FreedMem: 4096, Kills: 2, HighScore: 5}

	hud := toHUDState(tracker)

	assert.Equal(t, int64(4096), hud.FreedMem)
	assert.Equal(t, 2, hud.Kills)
	assert.Equal(t, 5, hud.HighScore)
}

// --- toStatusState ---

func TestToStatusStateMapsFields(t *testing.T) {
	gs := game.NewSession([]process.Info{process.NewInfo(1, "a", 0)}, game.Config{Speed: 2.0, TimeLimit: 30})
	gs.Start(80, 24)

	status := toStatusState(gs, 3)

	assert.Equal(t, 3, status.Alive)
	assert.Equal(t, 2.0, status.Speed)
	assert.Equal(t, 30, status.TimeLimit)
	assert.Nil(t, status.Confirming)
}

func TestToStatusStateIncludesConfirming(t *testing.T) {
	gs := game.NewSession([]process.Info{process.NewInfo(42, "suspect", 0)}, game.Config{Confirm: true, Speed: 1.0})
	gs.Start(80, 24)
	gs.RequestConfirm(gs.Targets()[0])

	status := toStatusState(gs, 1)

	require.NotNil(t, status.Confirming)
	assert.Equal(t, 42, status.Confirming.PID)
}
