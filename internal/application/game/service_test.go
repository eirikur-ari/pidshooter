package game

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// --- buildFrame ---

func TestBuildFrameAliveTargetIncluded(t *testing.T) {
	gs := game.NewSession([]process.Info{process.NewInfo(1, "a", 0)}, game.Config{Speed: 1.0})
	gs.Start(80, 24)

	f := buildFrame(gs, &score.Tracker{})

	require.Len(t, f.Targets, 1)
	assert.False(t, f.Targets[0].Killing)
}

func TestBuildFrameKillingTargetMarked(t *testing.T) {
	gs := game.NewSession([]process.Info{process.NewInfo(1, "a", 0)}, game.Config{Speed: 1.0})
	gs.Start(80, 24)
	gs.Targets()[0].Kill()

	f := buildFrame(gs, &score.Tracker{})

	require.Len(t, f.Targets, 1)
	assert.True(t, f.Targets[0].Killing)
}

func TestBuildFrameDeadTargetExcluded(t *testing.T) {
	gs := game.NewSession([]process.Info{process.NewInfo(1, "a", 0)}, game.Config{Speed: 1.0})
	gs.Start(80, 24)
	gs.Targets()[0].Kill()
	for i := 0; i < game.KillAnimationDuration; i++ {
		gs.Update(80, 24)
	}

	f := buildFrame(gs, &score.Tracker{})

	assert.Empty(t, f.Targets)
}

func TestBuildFrameHUDReflectsStats(t *testing.T) {
	gs := game.NewSession([]process.Info{process.NewInfo(1, "a", 4096)}, game.Config{Speed: 1.0})
	gs.Start(80, 24)
	tracker := &score.Tracker{}
	tracker.RecordKill(gs.Targets()[0].Rss)

	f := buildFrame(gs, tracker)

	assert.Equal(t, 1, f.HUD.Kills)
	assert.Equal(t, int64(4096), f.HUD.FreedMem)
}

func TestBuildFrameStatusBarAliveCount(t *testing.T) {
	processes := []process.Info{
		process.NewInfo(1, "a", 0),
		process.NewInfo(2, "b", 0),
	}
	gs := game.NewSession(processes, game.Config{Speed: 1.0})
	gs.Start(80, 24)
	gs.Targets()[0].Kill()

	f := buildFrame(gs, &score.Tracker{})

	assert.Equal(t, 1, f.StatusBar.Alive)
}

func TestBuildFrameNoConfirmPending(t *testing.T) {
	gs := game.NewSession(nil, game.Config{Speed: 1.0})
	gs.Start(80, 24)

	f := buildFrame(gs, &score.Tracker{})

	assert.Nil(t, f.StatusBar.Confirming)
}

func TestFindProcessesListError(t *testing.T) {
	svc := NewService(
		&fake.Process{ListErr: errors.New("ps failed")},
		&fake.Renderer{},
		fake.NewInputSource(),
	)
	_, err := svc.FindProcesses([]string{"foo"})
	require.Error(t, err)
}

func TestFindProcessesNoMatches(t *testing.T) {
	svc := NewService(&fake.Process{}, &fake.Renderer{}, fake.NewInputSource())

	processes, err := svc.FindProcesses([]string{"nonexistent"})

	assert.NoError(t, err)
	assert.Empty(t, processes)
}

func TestServiceApplyKillsCompletesPendingKill(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	svc := NewService(&fake.Process{}, &fake.Renderer{}, fake.NewInputSource())
	svc.kills = make(chan killSignal, 1)

	target := game.NewTarget(info, movement.NewBounds(80, 24))
	svc.kills <- killSignal{target: target}

	tracker := &score.Tracker{}
	svc.applyKills(tracker)

	assert.Equal(t, game.Killing, target.State)
	assert.Equal(t, 1, tracker.Kills)
}

func TestServiceApplyKillsReapsAlreadyKilledTarget(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	svc := NewService(&fake.Process{}, &fake.Renderer{}, fake.NewInputSource())
	svc.kills = make(chan killSignal, 1)

	target := game.NewTarget(info, movement.NewBounds(80, 24))
	svc.kills <- killSignal{target: target, shouldReap: true}

	tracker := &score.Tracker{}
	svc.applyKills(tracker)

	assert.Equal(t, game.Dead, target.State)
	assert.Equal(t, 0, tracker.Kills, "reaping an already-gone target should not award a kill")
}

func TestServiceApplyKillsEmptyChannelNoOps(t *testing.T) {
	svc := NewService(&fake.Process{}, &fake.Renderer{}, fake.NewInputSource())
	svc.kills = make(chan killSignal, 1)

	tracker := &score.Tracker{}
	svc.applyKills(tracker) // must not block

	assert.Equal(t, 0, tracker.Kills)
}

// --- kill ---

func TestServiceKillProtectedPIDReturnsError(t *testing.T) {
	fp := &fake.Process{}
	svc := NewService(fp, &fake.Renderer{}, fake.NewInputSource())
	target := game.NewTarget(process.NewInfo(1, "init", 0), movement.NewBounds(80, 24))

	killed, err := svc.kill(target)

	assert.False(t, killed)
	require.Error(t, err)
	assert.Empty(t, fp.KilledPIDs)
}

func TestServiceKillLookupErrorReturnsErrAlreadyKilled(t *testing.T) {
	fp := &fake.Process{LookupNameErr: errors.New("ps lookup failed")}
	svc := NewService(fp, &fake.Renderer{}, fake.NewInputSource())
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, err := svc.kill(target)

	assert.False(t, killed)
	require.Error(t, err)
	assert.ErrorIs(t, err, errAlreadyKilled)
	assert.ErrorContains(t, err, "ps lookup failed", "expected the underlying ps error to still be visible")
	assert.Empty(t, fp.KilledPIDs)
}

func TestServiceKillNameMismatchReturnsErrAlreadyKilled(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "somethingElse"}
	svc := NewService(fp, &fake.Renderer{}, fake.NewInputSource())
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, err := svc.kill(target)

	assert.False(t, killed)
	assert.ErrorIs(t, err, errAlreadyKilled)
	assert.Empty(t, fp.KilledPIDs)
}

func TestServiceKillNameMatchInvokesKill(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "target"}
	svc := NewService(fp, &fake.Renderer{}, fake.NewInputSource())
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, err := svc.kill(target)

	assert.True(t, killed)
	assert.NoError(t, err)
	assert.Equal(t, []int{100}, fp.KilledPIDs)
}
