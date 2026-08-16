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

func TestHandler_OnQuit_StopsGame(t *testing.T) {
	g := fixture.Game(nil, game.Config{})

	newHandler(g).OnQuit()

	assert.False(t, g.IsRunning())
}

func TestHandler_OnQuit_CancelsConfirmWhenPending(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	h := newHandler(g)

	h.OnQuit()

	assert.True(t, g.IsRunning(), "OnQuit should cancel confirm, not stop, when confirm is pending")
	assert.Nil(t, g.ConfirmTarget())
}

// --- OnYes ---

func TestHandler_OnYes_ReturnsConfirmedTarget(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	h := newHandler(g)

	result := h.OnYes()

	require.NotNil(t, result)
	assert.Nil(t, g.ConfirmTarget())
}

// --- OnNo ---

func TestHandler_OnNo_CancelsPending(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	h := newHandler(g)

	h.OnNo()

	assert.Nil(t, g.ConfirmTarget())
	assert.True(t, g.IsRunning())
}

// --- OnSpeedUp / OnSpeedDown ---

func TestHandler_OnSpeedUp_IncreasesSpeed(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 2.0})

	newHandler(g).OnSpeedUp()

	assert.Equal(t, 2.5, g.Speed())
}

func TestHandler_OnSpeedDown_DecreasesSpeed(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 2.0})

	newHandler(g).OnSpeedDown()

	assert.Equal(t, 1.5, g.Speed())
}

// --- OnClickAt ---

func TestHandler_OnClickAt_ReturnsTarget(t *testing.T) {
	g := fixture.Game(fixture.Processes(1), game.Config{Speed: 1.0})
	tgt := g.Targets()[0]

	result := newHandler(g).OnClickAt(int(math.Round(tgt.Position.X)), int(math.Round(tgt.Position.Y)))

	require.NotNil(t, result)
}

func TestHandler_OnClickAt_MissReturnsNil(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 1.0})

	result := newHandler(g).OnClickAt(0, 0)

	assert.Nil(t, result)
}

func TestHandler_OnClickAt_NoOpWhenAlreadyConfirming(t *testing.T) {
	g := fixture.PendingConfirmGame(2)
	h := newHandler(g)
	pending := g.ConfirmTarget()

	second := g.Targets()[1]
	result := h.OnClickAt(int(math.Round(second.Position.X)), int(math.Round(second.Position.Y)))

	assert.Nil(t, result, "expected no result when already confirming")
	assert.Equal(t, pending, g.ConfirmTarget(), "expected pending confirmation unchanged")
}