package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// newRunningGame returns a Game in the Running state with the given fields set.
func newRunningGame(fields *Game) *Game {
	fields.Start(0, 0)
	return fields
}

// --- HandleKey ---

func TestHandleKey_Quit(t *testing.T) {
	g := newRunningGame(&Game{})
	g.HandleKey('q')
	assert.False(t, g.IsRunning())
}

func TestHandleKey_QuitUppercase(t *testing.T) {
	g := newRunningGame(&Game{})
	g.HandleKey('Q')
	assert.False(t, g.IsRunning())
}

func TestHandleKey_SpeedUp(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(2.0)})
	g.HandleKey('+')
	assert.Equal(t, 2.5, g.velocity.Speed())
}

func TestHandleKey_SpeedDown(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(2.0)})
	g.HandleKey('-')
	assert.Equal(t, 1.5, g.velocity.Speed())
}

func TestHandleKey_SpeedUpAlias(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(2.0)})
	g.HandleKey('=')
	assert.Equal(t, 2.5, g.velocity.Speed(), "expected speed=2.5 with '=' alias")
}

func TestHandleKey_SpeedDownAlias(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(2.0)})
	g.HandleKey('_')
	assert.Equal(t, 1.5, g.velocity.Speed(), "expected speed=1.5 with '_' alias")
}

func TestHandleKey_SpeedCapsAtMax(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(4.8)})
	g.HandleKey('+')
	assert.Equal(t, 5.0, g.velocity.Speed(), "expected speed capped at 5.0")
	g.HandleKey('+')
	assert.Equal(t, 5.0, g.velocity.Speed(), "expected speed still 5.0")
}

func TestHandleKey_SpeedCapsAtMin(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(0.3)})
	g.HandleKey('-')
	assert.Equal(t, 0.1, g.velocity.Speed(), "expected speed capped at 0.1")
}

func TestHandleKey_ConfirmYes_ReturnsTarget(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 4096}, State: Alive}
	g := newRunningGame(&Game{confirm: Confirmation{target: tgt}})

	target := g.HandleKey('y')

	require.NotNil(t, target)
	assert.Equal(t, tgt, target)
	assert.False(t, g.confirm.Pending())
}

func TestHandleKey_ConfirmNo(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 0}, State: Alive}
	g := newRunningGame(&Game{confirm: Confirmation{target: tgt}})

	target := g.HandleKey('n')

	assert.Nil(t, target)
	assert.False(t, g.confirm.Pending())
	assert.Equal(t, Alive, tgt.State)
}

func TestHandleKey_QCancelsConfirm(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 0}, State: Alive}
	g := newRunningGame(&Game{confirm: Confirmation{target: tgt}})

	g.HandleKey('q')

	assert.False(t, g.confirm.Pending())
	assert.True(t, g.IsRunning(), "q cancels confirm, doesn't quit")
}

// --- HandleClick ---

func TestHandleClick_ReturnsTarget(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: FrameVector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{tgt}}

	target := g.HandleClick(10, 5)

	require.NotNil(t, target)
	assert.Equal(t, tgt, target)
}

func TestHandleClick_SetsConfirmingInConfirmMode(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: FrameVector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{tgt}, confirm: NewConfirmation(true)}

	target := g.HandleClick(10, 5)

	assert.Nil(t, target, "expected no target in confirm mode (should set pending instead)")
	assert.Equal(t, tgt, g.confirm.target)
	assert.Equal(t, Alive, tgt.State)
}

func TestHandleClick_NoOpWhenAlreadyConfirming(t *testing.T) {
	existing := &Target{Info: process.Info{Pid: 1, Name: "other", Rss: 0}, State: Alive}
	tgt := &Target{Info: process.Info{Pid: 2, Name: "target", Rss: 1024}, Position: FrameVector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{tgt}, confirm: Confirmation{target: existing}}

	result := g.HandleClick(10, 5)

	assert.Nil(t, result, "expected no target when already confirming")
	assert.Equal(t, existing, g.confirm.target, "expected pending confirmation to remain unchanged")
}

func TestHandleClick_NoOpOnMiss(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: FrameVector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{tgt}}

	target := g.HandleClick(0, 0)

	assert.Nil(t, target)
	assert.Equal(t, Alive, tgt.State)
}
