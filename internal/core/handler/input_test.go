package handler_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/handler"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func newHandler(g *game.Game) handler.InputHandler {
	return handler.NewHandler(g)
}

// --- OnQuit ---

func TestHandler_OnQuit_StopsGame(t *testing.T) {
	g := game.New(nil, game.Config{})
	g.Start(80, 24)

	newHandler(g).OnQuit()

	assert.False(t, g.IsRunning())
}

func TestHandler_OnQuit_CancelsConfirmWhenPending(t *testing.T) {
	processes := []process.Info{process.NewInfo(1, "a", 0)}
	g := game.New(processes, game.Config{Confirm: true, Speed: 1.0})
	g.Start(80, 24)
	h := newHandler(g)
	tgt := g.Targets()[0]
	h.OnClickAt(int(math.Round(tgt.Position.X)), int(math.Round(tgt.Position.Y)))
	require.NotNil(t, g.ConfirmTarget(), "setup: expected confirm pending")

	h.OnQuit()

	assert.True(t, g.IsRunning(), "OnQuit should cancel confirm, not stop, when confirm is pending")
	assert.Nil(t, g.ConfirmTarget())
}

// --- OnYes ---

func TestHandler_OnYes_ReturnsConfirmedTarget(t *testing.T) {
	processes := []process.Info{process.NewInfo(1, "a", 0)}
	g := game.New(processes, game.Config{Confirm: true, Speed: 1.0})
	g.Start(80, 24)
	h := newHandler(g)
	tgt := g.Targets()[0]
	h.OnClickAt(int(math.Round(tgt.Position.X)), int(math.Round(tgt.Position.Y)))
	require.NotNil(t, g.ConfirmTarget(), "setup: expected confirm pending")

	result := h.OnYes()

	require.NotNil(t, result)
	assert.Nil(t, g.ConfirmTarget())
}

// --- OnNo ---

func TestHandler_OnNo_CancelsPending(t *testing.T) {
	processes := []process.Info{process.NewInfo(1, "a", 0)}
	g := game.New(processes, game.Config{Confirm: true, Speed: 1.0})
	g.Start(80, 24)
	h := newHandler(g)
	tgt := g.Targets()[0]
	h.OnClickAt(int(math.Round(tgt.Position.X)), int(math.Round(tgt.Position.Y)))
	require.NotNil(t, g.ConfirmTarget(), "setup: expected confirm pending")

	h.OnNo()

	assert.Nil(t, g.ConfirmTarget())
	assert.True(t, g.IsRunning())
}

// --- OnSpeedUp / OnSpeedDown ---

func TestHandler_OnSpeedUp_IncreasesSpeed(t *testing.T) {
	g := game.New(nil, game.Config{Speed: 2.0})
	g.Start(80, 24)

	newHandler(g).OnSpeedUp()

	assert.Equal(t, 2.5, g.Speed())
}

func TestHandler_OnSpeedDown_DecreasesSpeed(t *testing.T) {
	g := game.New(nil, game.Config{Speed: 2.0})
	g.Start(80, 24)

	newHandler(g).OnSpeedDown()

	assert.Equal(t, 1.5, g.Speed())
}

// --- OnClickAt ---

func TestHandler_OnClickAt_ReturnsTarget(t *testing.T) {
	processes := []process.Info{process.NewInfo(1, "a", 0)}
	g := game.New(processes, game.Config{Speed: 1.0})
	g.Start(80, 24)
	tgt := g.Targets()[0]

	result := newHandler(g).OnClickAt(int(math.Round(tgt.Position.X)), int(math.Round(tgt.Position.Y)))

	require.NotNil(t, result)
}

func TestHandler_OnClickAt_MissReturnsNil(t *testing.T) {
	g := game.New(nil, game.Config{Speed: 1.0})
	g.Start(80, 24)

	result := newHandler(g).OnClickAt(0, 0)

	assert.Nil(t, result)
}

func TestHandler_OnClickAt_NoOpWhenAlreadyConfirming(t *testing.T) {
	processes := []process.Info{
		process.NewInfo(1, "a", 0),
		process.NewInfo(2, "b", 0),
	}
	g := game.New(processes, game.Config{Confirm: true, Speed: 1.0})
	g.Start(80, 24)
	h := newHandler(g)
	first := g.Targets()[0]
	h.OnClickAt(int(math.Round(first.Position.X)), int(math.Round(first.Position.Y)))
	require.NotNil(t, g.ConfirmTarget(), "setup: expected confirm pending")
	pending := g.ConfirmTarget()

	second := g.Targets()[1]
	result := h.OnClickAt(int(math.Round(second.Position.X)), int(math.Round(second.Position.Y)))

	assert.Nil(t, result, "expected no result when already confirming")
	assert.Equal(t, pending, g.ConfirmTarget(), "expected pending confirmation unchanged")
}
