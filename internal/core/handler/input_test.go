package handler_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/handler"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fixture"
)

func newHandler(g *game.Game) *handler.Handler {
	return handler.NewHandler(g)
}

// --- OnQuit ---

func TestHandlerOnQuitStopsGame(t *testing.T) {
	g := fixture.Game(nil, game.Config{})

	newHandler(g).OnQuit()

	assert.False(t, g.IsRunning())
}

func TestHandlerOnQuitCancelsConfirmWhenPending(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	h := newHandler(g)

	h.OnQuit()

	assert.True(t, g.IsRunning(), "OnQuit should cancel confirm, not stop, when confirm is pending")
	assert.Nil(t, g.ConfirmTarget())
}

// --- OnYes ---

func TestHandlerOnYesReturnsConfirmedTarget(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	h := newHandler(g)

	result := h.OnYes()

	require.NotNil(t, result)
	assert.Nil(t, g.ConfirmTarget())
}

// --- OnNo ---

func TestHandlerOnNoCancelsPending(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	h := newHandler(g)

	h.OnNo()

	assert.Nil(t, g.ConfirmTarget())
	assert.True(t, g.IsRunning())
}

// --- OnSpeedUp / OnSpeedDown ---

func TestHandlerOnSpeedUpIncreasesSpeed(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 2.0})

	newHandler(g).OnSpeedUp()

	assert.Equal(t, 2.5, g.Speed())
}

func TestHandlerOnSpeedDownDecreasesSpeed(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 2.0})

	newHandler(g).OnSpeedDown()

	assert.Equal(t, 1.5, g.Speed())
}

// --- OnClickAt ---

func TestHandlerOnClickAtReturnsTarget(t *testing.T) {
	g := fixture.Game(fixture.Processes(1), game.Config{Speed: 1.0})
	tgt := g.Targets()[0]

	result := newHandler(g).OnClickAt(int(math.Round(tgt.Position.X)), int(math.Round(tgt.Position.Y)))

	require.NotNil(t, result)
}

func TestHandlerOnClickAtMissReturnsNil(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 1.0})

	result := newHandler(g).OnClickAt(0, 0)

	assert.Nil(t, result)
}

func TestHandlerOnClickAtNoOpWhenAlreadyConfirming(t *testing.T) {
	g := fixture.PendingConfirmGame(2)
	h := newHandler(g)
	pending := g.ConfirmTarget()

	second := g.Targets()[1]
	result := h.OnClickAt(int(math.Round(second.Position.X)), int(math.Round(second.Position.Y)))

	assert.Nil(t, result, "expected no result when already confirming")
	assert.Equal(t, pending, g.ConfirmTarget(), "expected pending confirmation unchanged")
}