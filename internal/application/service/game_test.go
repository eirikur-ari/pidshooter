package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// --- toConfirmViewState ---

func TestToConfirmViewState_NilInput(t *testing.T) {
	assert.Nil(t, toConfirmViewState(nil))
}

func TestToConfirmViewState_MapsFields(t *testing.T) {
	tgt := game.NewTarget(process.Info{Pid: 42, Name: "suspect"}, game.FrameBounds{Width: 80, Height: 24})
	vs := toConfirmViewState(tgt)
	require.NotNil(t, vs)
	assert.Equal(t, 42, vs.PID)
	assert.Equal(t, "suspect", vs.Name)
}

// --- buildFrame ---

func TestBuildFrame_AliveTargetIncluded(t *testing.T) {
	g := game.New([]process.Info{{Pid: 1, Name: "a", Rss: 0}}, game.Config{Speed: 1.0})
	g.Start(80, 24)

	f := buildFrame(g)

	require.Len(t, f.Targets, 1)
	assert.False(t, f.Targets[0].Killing)
}

func TestBuildFrame_KillingTargetMarked(t *testing.T) {
	g := game.New([]process.Info{{Pid: 1, Name: "a", Rss: 0}}, game.Config{Speed: 1.0})
	g.Start(80, 24)
	g.Kill(g.Targets()[0])

	f := buildFrame(g)

	require.Len(t, f.Targets, 1)
	assert.True(t, f.Targets[0].Killing)
}

func TestBuildFrame_DeadTargetExcluded(t *testing.T) {
	g := game.New([]process.Info{{Pid: 1, Name: "a", Rss: 0}}, game.Config{Speed: 1.0})
	g.Start(80, 24)
	g.Kill(g.Targets()[0])
	for i := 0; i < game.KillAnimationDuration; i++ {
		g.Update(80, 24)
	}

	f := buildFrame(g)

	assert.Empty(t, f.Targets)
}

func TestBuildFrame_HUDReflectsStats(t *testing.T) {
	g := game.New([]process.Info{{Pid: 1, Name: "a", Rss: 4096}}, game.Config{Speed: 1.0})
	g.Start(80, 24)
	g.Kill(g.Targets()[0])

	f := buildFrame(g)

	assert.Equal(t, 1, f.HUD.Kills)
	assert.Equal(t, int64(4096), f.HUD.FreedMem)
}

func TestBuildFrame_StatusBarAliveCount(t *testing.T) {
	processes := []process.Info{
		{Pid: 1, Name: "a", Rss: 0},
		{Pid: 2, Name: "b", Rss: 0},
	}
	g := game.New(processes, game.Config{Speed: 1.0})
	g.Start(80, 24)
	g.Kill(g.Targets()[0])

	f := buildFrame(g)

	assert.Equal(t, 1, f.StatusBar.Alive)
}

func TestBuildFrame_NoConfirmPending(t *testing.T) {
	g := game.New(nil, game.Config{Speed: 1.0})
	g.Start(80, 24)

	f := buildFrame(g)

	assert.Nil(t, f.StatusBar.Confirming)
}

func TestGameService_FinderError(t *testing.T) {
	svc := NewGameService(
		&fake.Process{FindErr: errors.New("ps failed")},
		&fake.Store{},
		&fake.Renderer{},
		fake.NewInputSource(),
	)
	err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"foo"}, Speed: 2.0, TimeLimit: 30})
	require.Error(t, err)
}

func TestGameService_ApplyKills_CompletesPendingKill(t *testing.T) {
	info := process.Info{Pid: 100, Name: "target", Rss: 4096}
	svc := NewGameService(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource())
	svc.kills = make(chan *game.Target, 1)

	g := game.New([]process.Info{info}, game.Config{Speed: 2.0})

	target := game.NewTarget(info, game.FrameBounds{Width: 80, Height: 24})
	svc.kills <- target

	svc.applyKills(g)

	assert.Equal(t, game.Killing, target.State)
	assert.Equal(t, 1, g.Kills())
}

func TestGameService_ApplyKills_EmptyChannelNoOps(t *testing.T) {
	svc := NewGameService(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource())
	svc.kills = make(chan *game.Target, 1)

	g := game.New([]process.Info{}, game.Config{Speed: 2.0})

	svc.applyKills(g) // must not block

	assert.Equal(t, 0, g.Kills())
}

func TestGameService_NoProcesses(t *testing.T) {
	svc := NewGameService(
		&fake.Process{},
		&fake.Store{},
		&fake.Renderer{},
		fake.NewInputSource(),
	)
	err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"nonexistent"}, Speed: 2.0, TimeLimit: 30})
	assert.NoError(t, err)
}
