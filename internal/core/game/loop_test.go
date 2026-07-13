package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestKill_TransitionsToKilling(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 2048}, State: Alive}
	g := &Game{}

	g.Kill(tgt)

	assert.Equal(t, Killing, tgt.State)
	assert.Equal(t, 1, g.kills)
	assert.Equal(t, int64(2048), g.freedMem)
}

func TestKill_NoOpWhenNotAlive(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 2048}, State: Dead}
	g := &Game{}

	g.Kill(tgt)

	assert.Equal(t, 0, g.kills)
}

func TestUpdate_StopsWhenTimeLimitExpired(t *testing.T) {
	g := New(nil, Config{TimeLimit: 1})
	g.Start(0, 0)
	g.timer.start = time.Now().Add(-2 * time.Second)

	g.Update(80, 24)

	assert.False(t, g.IsRunning())
}

func TestUpdate_StopsWhenAllTargetsDead(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 0}, State: Dead}
	g := &Game{targets: []*Target{tgt}}
	g.Start(0, 0)

	g.Update(80, 24)

	assert.False(t, g.IsRunning())
}

func TestFrame_AliveTargetIncluded(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "myapp", Rss: 1024}, Position: FrameVector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{tgt}}

	frame := g.Frame()

	require.Len(t, frame.Targets, 1)
	tv := frame.Targets[0]
	assert.Equal(t, 10, tv.X)
	assert.Equal(t, 5, tv.Y)
	assert.False(t, tv.Killing)
}

func TestFrame_DeadTargetExcluded(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "myapp", Rss: 1024}, State: Dead}
	g := &Game{targets: []*Target{tgt}}

	frame := g.Frame()

	assert.Empty(t, frame.Targets)
}

func TestFrame_KillingTargetMarked(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "myapp", Rss: 1024}, State: Killing}
	g := &Game{targets: []*Target{tgt}}

	frame := g.Frame()

	require.Len(t, frame.Targets, 1)
	assert.True(t, frame.Targets[0].Killing)
}

func TestFrame_HUDReflectsStats(t *testing.T) {
	g := &Game{targets: []*Target{}}
	g.kills = 3
	g.freedMem = 2048
	g.highScore = 10

	frame := g.Frame()

	assert.Equal(t, 3, frame.HUD.Kills)
	assert.Equal(t, int64(2048), frame.HUD.FreedMem)
	assert.Equal(t, 10, frame.HUD.HighScore)
}

func TestFrame_ConfirmStateInStatusBar(t *testing.T) {
	target := &Target{Info: process.Info{Pid: 42, Name: "suspect", Rss: 0}, State: Alive}
	g := &Game{targets: []*Target{target}, confirm: Confirmation{target: target}}

	frame := g.Frame()

	cs := frame.StatusBar.Confirming
	require.NotNil(t, cs, "expected ConfirmState in status bar")
	assert.Equal(t, 42, cs.PID)
	assert.Equal(t, "suspect", cs.Name)
}

func TestFrame_StatusBarAliveCount(t *testing.T) {
	alive := &Target{Info: process.Info{Pid: 1, Name: "a", Rss: 0}, State: Alive}
	dead := &Target{Info: process.Info{Pid: 2, Name: "b", Rss: 0}, State: Dead}
	killing := &Target{Info: process.Info{Pid: 3, Name: "c", Rss: 0}, State: Killing}
	g := &Game{targets: []*Target{alive, dead, killing}}

	frame := g.Frame()

	assert.Equal(t, 1, frame.StatusBar.Alive)
}
