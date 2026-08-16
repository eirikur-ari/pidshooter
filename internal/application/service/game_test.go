package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

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

// --- buildFrame ---

func TestBuildFrameAliveTargetIncluded(t *testing.T) {
	g := game.New([]process.Info{process.NewInfo(1, "a", 0)}, game.Config{Speed: 1.0})
	g.Start(80, 24)

	f := buildFrame(g)

	require.Len(t, f.Targets, 1)
	assert.False(t, f.Targets[0].Killing)
}

func TestBuildFrameKillingTargetMarked(t *testing.T) {
	g := game.New([]process.Info{process.NewInfo(1, "a", 0)}, game.Config{Speed: 1.0})
	g.Start(80, 24)
	g.Kill(g.Targets()[0])

	f := buildFrame(g)

	require.Len(t, f.Targets, 1)
	assert.True(t, f.Targets[0].Killing)
}

func TestBuildFrameDeadTargetExcluded(t *testing.T) {
	g := game.New([]process.Info{process.NewInfo(1, "a", 0)}, game.Config{Speed: 1.0})
	g.Start(80, 24)
	g.Kill(g.Targets()[0])
	for i := 0; i < game.KillAnimationDuration; i++ {
		g.Update(80, 24)
	}

	f := buildFrame(g)

	assert.Empty(t, f.Targets)
}

func TestBuildFrameHUDReflectsStats(t *testing.T) {
	g := game.New([]process.Info{process.NewInfo(1, "a", 4096)}, game.Config{Speed: 1.0})
	g.Start(80, 24)
	g.Kill(g.Targets()[0])

	f := buildFrame(g)

	assert.Equal(t, 1, f.HUD.Kills)
	assert.Equal(t, int64(4096), f.HUD.FreedMem)
}

func TestBuildFrameStatusBarAliveCount(t *testing.T) {
	processes := []process.Info{
		process.NewInfo(1, "a", 0),
		process.NewInfo(2, "b", 0),
	}
	g := game.New(processes, game.Config{Speed: 1.0})
	g.Start(80, 24)
	g.Kill(g.Targets()[0])

	f := buildFrame(g)

	assert.Equal(t, 1, f.StatusBar.Alive)
}

func TestBuildFrameNoConfirmPending(t *testing.T) {
	g := game.New(nil, game.Config{Speed: 1.0})
	g.Start(80, 24)

	f := buildFrame(g)

	assert.Nil(t, f.StatusBar.Confirming)
}

func TestGameServiceFinderError(t *testing.T) {
	svc := NewGameService(
		&fake.Process{ListErr: errors.New("ps failed")},
		&fake.Store{},
		&fake.Renderer{},
		fake.NewInputSource(),
	)
	err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"foo"}, Speed: 2.0, TimeLimit: 30})
	require.Error(t, err)
}

func TestGameServiceApplyKillsCompletesPendingKill(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	svc := NewGameService(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource())
	svc.kills = make(chan *game.Target, 1)

	g := game.New([]process.Info{info}, game.Config{Speed: 2.0})

	target := game.NewTarget(info, movement.NewBounds(80, 24))
	svc.kills <- target

	svc.applyKills(g)

	assert.Equal(t, game.Killing, target.State)
	assert.Equal(t, 1, g.Kills())
}

func TestGameServiceApplyKillsEmptyChannelNoOps(t *testing.T) {
	svc := NewGameService(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource())
	svc.kills = make(chan *game.Target, 1)

	g := game.New([]process.Info{}, game.Config{Speed: 2.0})

	svc.applyKills(g) // must not block

	assert.Equal(t, 0, g.Kills())
}

// --- kill ---

func TestGameServiceKillProtectedPIDReturnsError(t *testing.T) {
	fp := &fake.Process{}
	svc := NewGameService(fp, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource())
	target := game.NewTarget(process.NewInfo(1, "init", 0), movement.NewBounds(80, 24))

	killed, err := svc.kill(target)

	assert.False(t, killed)
	require.Error(t, err)
	assert.Empty(t, fp.KilledPIDs)
}

func TestGameServiceKillLookupErrorReturnsError(t *testing.T) {
	fp := &fake.Process{LookupNameErr: errors.New("ps lookup failed")}
	svc := NewGameService(fp, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource())
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, err := svc.kill(target)

	assert.False(t, killed)
	require.Error(t, err)
	assert.Empty(t, fp.KilledPIDs)
}

func TestGameServiceKillNameMismatchSkipsKillWithoutError(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "somethingElse"}
	svc := NewGameService(fp, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource())
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, err := svc.kill(target)

	assert.False(t, killed)
	assert.NoError(t, err)
	assert.Empty(t, fp.KilledPIDs)
}

func TestGameServiceKillNameMatchInvokesKill(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "target"}
	svc := NewGameService(fp, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource())
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, err := svc.kill(target)

	assert.True(t, killed)
	assert.NoError(t, err)
	assert.Equal(t, []int{100}, fp.KilledPIDs)
}

func TestGameServiceNoProcesses(t *testing.T) {
	svc := NewGameService(
		&fake.Process{},
		&fake.Store{},
		&fake.Renderer{},
		fake.NewInputSource(),
	)
	err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"nonexistent"}, Speed: 2.0, TimeLimit: 30})
	assert.NoError(t, err)
}
