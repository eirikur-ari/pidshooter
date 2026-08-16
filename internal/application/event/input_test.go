package event

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/handler"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func newDispatcher(g *game.Game) *Dispatcher {
	return NewDispatcher(handler.NewHandler(g))
}

// --- ClickEvent ---

func TestDispatcher_Click_ReturnsHitTarget(t *testing.T) {
	processes := []process.Info{process.NewInfo(1, "a", 0)}
	g := game.New(processes, game.Config{Speed: 1.0})
	g.Start(80, 24)
	d := newDispatcher(g)
	tgt := g.Targets()[0]

	result := d.Dispatch(outbound.ClickEvent{X: int(math.Round(tgt.Position.X)), Y: int(math.Round(tgt.Position.Y))})

	assert.Equal(t, tgt, result)
}

func TestDispatcher_Click_MissReturnsNil(t *testing.T) {
	g := game.New(nil, game.Config{Speed: 1.0})
	g.Start(80, 24)
	d := newDispatcher(g)

	result := d.Dispatch(outbound.ClickEvent{X: 0, Y: 0})

	assert.Nil(t, result)
}

// --- KeyCode ---

func TestDispatcher_Escape_StopsGame(t *testing.T) {
	g := game.New(nil, game.Config{})
	g.Start(80, 24)

	newDispatcher(g).Dispatch(outbound.KeyEvent{Key: outbound.KeyEscape})

	assert.False(t, g.IsRunning())
}

func TestDispatcher_CtrlC_StopsGame(t *testing.T) {
	g := game.New(nil, game.Config{})
	g.Start(80, 24)

	newDispatcher(g).Dispatch(outbound.KeyEvent{Key: outbound.KeyCtrlC})

	assert.False(t, g.IsRunning())
}

func TestDispatcher_CtrlZ_StopsGame(t *testing.T) {
	g := game.New(nil, game.Config{})
	g.Start(80, 24)

	newDispatcher(g).Dispatch(outbound.KeyEvent{Key: outbound.KeyCtrlZ})

	assert.False(t, g.IsRunning())
}

// --- Rune: quit ---

func TestDispatcher_Q_StopsGame(t *testing.T) {
	g := game.New(nil, game.Config{})
	g.Start(80, 24)

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: 'q'})

	assert.False(t, g.IsRunning())
}

func TestDispatcher_QUppercase_StopsGame(t *testing.T) {
	g := game.New(nil, game.Config{})
	g.Start(80, 24)

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: 'Q'})

	assert.False(t, g.IsRunning())
}

// --- Rune: confirm yes ---

func TestDispatcher_Y_ReturnsConfirmedTarget(t *testing.T) {
	processes := []process.Info{process.NewInfo(1, "a", 0)}
	g := game.New(processes, game.Config{Confirm: true, Speed: 1.0})
	g.Start(80, 24)
	d := newDispatcher(g)
	tgt := g.Targets()[0]
	d.Dispatch(outbound.ClickEvent{X: int(math.Round(tgt.Position.X)), Y: int(math.Round(tgt.Position.Y))})
	require.NotNil(t, g.ConfirmTarget(), "setup: expected confirm pending")

	result := d.Dispatch(outbound.KeyEvent{Ch: 'y'})

	assert.Equal(t, tgt, result)
	assert.Nil(t, g.ConfirmTarget())
}

func TestDispatcher_YUppercase_AcceptsConfirmation(t *testing.T) {
	processes := []process.Info{process.NewInfo(1, "a", 0)}
	g := game.New(processes, game.Config{Confirm: true, Speed: 1.0})
	g.Start(80, 24)
	d := newDispatcher(g)
	tgt := g.Targets()[0]
	d.Dispatch(outbound.ClickEvent{X: int(math.Round(tgt.Position.X)), Y: int(math.Round(tgt.Position.Y))})
	require.NotNil(t, g.ConfirmTarget(), "setup: expected confirm pending")

	result := d.Dispatch(outbound.KeyEvent{Ch: 'Y'})

	assert.Equal(t, tgt, result)
	assert.Nil(t, g.ConfirmTarget())
}

// --- Rune: confirm no ---

func TestDispatcher_N_CancelsConfirmation(t *testing.T) {
	processes := []process.Info{process.NewInfo(1, "a", 0)}
	g := game.New(processes, game.Config{Confirm: true, Speed: 1.0})
	g.Start(80, 24)
	d := newDispatcher(g)
	tgt := g.Targets()[0]
	d.Dispatch(outbound.ClickEvent{X: int(math.Round(tgt.Position.X)), Y: int(math.Round(tgt.Position.Y))})
	require.NotNil(t, g.ConfirmTarget(), "setup: expected confirm pending")

	d.Dispatch(outbound.KeyEvent{Ch: 'n'})

	assert.Nil(t, g.ConfirmTarget())
}

func TestDispatcher_NUppercase_CancelsConfirmation(t *testing.T) {
	processes := []process.Info{process.NewInfo(1, "a", 0)}
	g := game.New(processes, game.Config{Confirm: true, Speed: 1.0})
	g.Start(80, 24)
	d := newDispatcher(g)
	tgt := g.Targets()[0]
	d.Dispatch(outbound.ClickEvent{X: int(math.Round(tgt.Position.X)), Y: int(math.Round(tgt.Position.Y))})
	require.NotNil(t, g.ConfirmTarget(), "setup: expected confirm pending")

	d.Dispatch(outbound.KeyEvent{Ch: 'N'})

	assert.Nil(t, g.ConfirmTarget())
}

// --- Rune: speed ---

func TestDispatcher_Plus_IncreasesSpeed(t *testing.T) {
	g := game.New(nil, game.Config{Speed: 2.0})
	g.Start(80, 24)

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: '+'})

	assert.Equal(t, 2.5, g.Speed())
}

func TestDispatcher_Equals_IncreasesSpeed(t *testing.T) {
	g := game.New(nil, game.Config{Speed: 2.0})
	g.Start(80, 24)

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: '='})

	assert.Equal(t, 2.5, g.Speed())
}

func TestDispatcher_Minus_DecreasesSpeed(t *testing.T) {
	g := game.New(nil, game.Config{Speed: 2.0})
	g.Start(80, 24)

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: '-'})

	assert.Equal(t, 1.5, g.Speed())
}

func TestDispatcher_Underscore_DecreasesSpeed(t *testing.T) {
	g := game.New(nil, game.Config{Speed: 2.0})
	g.Start(80, 24)

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: '_'})

	assert.Equal(t, 1.5, g.Speed())
}

// --- Unknown ---

func TestDispatcher_UnknownRune_NoOp(t *testing.T) {
	g := game.New(nil, game.Config{Speed: 2.0})
	g.Start(80, 24)
	d := newDispatcher(g)

	result := d.Dispatch(outbound.KeyEvent{Ch: 'z'})

	assert.Nil(t, result)
	assert.True(t, g.IsRunning())
	assert.Equal(t, 2.0, g.Speed())
}