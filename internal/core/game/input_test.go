package game_test

import (
	"math"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil/fixture"
)

func newInput(g *game.Game) *game.Input {
	return game.NewInput(g)
}

// --- OnQuit ---

func TestInputOnQuitStopsGame(t *testing.T) {
	g := fixture.Game(nil, game.Config{})

	newInput(g).OnQuit()

	assert.False(t, g.IsRunning())
}

func TestInputOnQuitCancelsConfirmWhenPending(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	in := newInput(g)

	in.OnQuit()

	assert.True(t, g.IsRunning(), "OnQuit should cancel confirm, not stop, when confirm is pending")
	assert.Nil(t, g.ConfirmTarget())
}

// --- OnYes ---

func TestInputOnYesReturnsConfirmedTarget(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	in := newInput(g)

	result := in.OnYes()

	require.NotNil(t, result)
	assert.Nil(t, g.ConfirmTarget())
}

// --- OnNo ---

func TestInputOnNoCancelsPending(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	in := newInput(g)

	in.OnNo()

	assert.Nil(t, g.ConfirmTarget())
	assert.True(t, g.IsRunning())
}

// --- OnSpeedUp / OnSpeedDown ---

func TestInputOnSpeedUpIncreasesSpeed(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 2.0})

	newInput(g).OnSpeedUp()

	assert.Equal(t, 2.5, g.Throttle().Speed())
}

func TestInputOnSpeedDownDecreasesSpeed(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 2.0})

	newInput(g).OnSpeedDown()

	assert.Equal(t, 1.5, g.Throttle().Speed())
}

// --- OnClickAt ---

func TestInputOnClickAtReturnsTarget(t *testing.T) {
	g := fixture.Game(fixture.Processes(1), game.Config{Speed: 1.0})
	tgt := g.Targets()[0]

	result := newInput(g).OnClickAt(int(math.Round(tgt.Position.X)), int(math.Round(tgt.Position.Y)))

	require.NotNil(t, result)
}

func TestInputOnClickAtMissReturnsNil(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 1.0})

	result := newInput(g).OnClickAt(0, 0)

	assert.Nil(t, result)
}

func TestInputOnClickAtNoOpWhenAlreadyConfirming(t *testing.T) {
	g := fixture.PendingConfirmGame(2)
	in := newInput(g)
	pending := g.ConfirmTarget()

	second := g.Targets()[1]
	result := in.OnClickAt(int(math.Round(second.Position.X)), int(math.Round(second.Position.Y)))

	assert.Nil(t, result, "expected no result when already confirming")
	assert.Equal(t, pending, g.ConfirmTarget(), "expected pending confirmation unchanged")
}
