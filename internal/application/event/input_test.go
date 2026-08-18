package event

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fixture"
)

func newDispatcher(g *game.Game) *Dispatcher {
	return NewDispatcher(game.NewInput(g))
}

// --- ClickEvent ---

func TestDispatcherClickReturnsHitTarget(t *testing.T) {
	g := fixture.Game(fixture.Processes(1), game.Config{Speed: 1.0})
	d := newDispatcher(g)
	tgt := g.Targets()[0]

	result := d.Dispatch(outbound.ClickEvent{X: int(math.Round(tgt.Position.X)), Y: int(math.Round(tgt.Position.Y))})

	assert.Equal(t, tgt, result)
}

func TestDispatcherClickMissReturnsNil(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 1.0})
	d := newDispatcher(g)

	result := d.Dispatch(outbound.ClickEvent{X: 0, Y: 0})

	assert.Nil(t, result)
}

// --- KeyCode ---

func TestDispatcherEscapeStopsGame(t *testing.T) {
	g := fixture.Game(nil, game.Config{})

	newDispatcher(g).Dispatch(outbound.KeyEvent{Key: outbound.KeyEscape})

	assert.False(t, g.IsRunning())
}

func TestDispatcherCtrlCStopsGame(t *testing.T) {
	g := fixture.Game(nil, game.Config{})

	newDispatcher(g).Dispatch(outbound.KeyEvent{Key: outbound.KeyCtrlC})

	assert.False(t, g.IsRunning())
}

func TestDispatcherCtrlZStopsGame(t *testing.T) {
	g := fixture.Game(nil, game.Config{})

	newDispatcher(g).Dispatch(outbound.KeyEvent{Key: outbound.KeyCtrlZ})

	assert.False(t, g.IsRunning())
}

// --- Rune: quit ---

func TestDispatcherQStopsGame(t *testing.T) {
	g := fixture.Game(nil, game.Config{})

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: 'q'})

	assert.False(t, g.IsRunning())
}

func TestDispatcherQUppercaseStopsGame(t *testing.T) {
	g := fixture.Game(nil, game.Config{})

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: 'Q'})

	assert.False(t, g.IsRunning())
}

// --- Rune: confirm yes ---

func TestDispatcherYReturnsConfirmedTarget(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	d := newDispatcher(g)
	tgt := g.Targets()[0]

	result := d.Dispatch(outbound.KeyEvent{Ch: 'y'})

	assert.Equal(t, tgt, result)
	assert.Nil(t, g.PendingConfirm())
}

func TestDispatcherYUppercaseAcceptsConfirmation(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	d := newDispatcher(g)
	tgt := g.Targets()[0]

	result := d.Dispatch(outbound.KeyEvent{Ch: 'Y'})

	assert.Equal(t, tgt, result)
	assert.Nil(t, g.PendingConfirm())
}

// --- Rune: confirm no ---

func TestDispatcherNCancelsConfirmation(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	d := newDispatcher(g)

	d.Dispatch(outbound.KeyEvent{Ch: 'n'})

	assert.Nil(t, g.PendingConfirm())
}

func TestDispatcherNUppercaseCancelsConfirmation(t *testing.T) {
	g := fixture.PendingConfirmGame(1)
	d := newDispatcher(g)

	d.Dispatch(outbound.KeyEvent{Ch: 'N'})

	assert.Nil(t, g.PendingConfirm())
}

// --- Rune: speed ---

func TestDispatcherPlusIncreasesSpeed(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 2.0})

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: '+'})

	assert.Equal(t, 2.5, g.Throttle().Speed())
}

func TestDispatcherEqualsIncreasesSpeed(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 2.0})

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: '='})

	assert.Equal(t, 2.5, g.Throttle().Speed())
}

func TestDispatcherMinusDecreasesSpeed(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 2.0})

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: '-'})

	assert.Equal(t, 1.5, g.Throttle().Speed())
}

func TestDispatcherUnderscoreDecreasesSpeed(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 2.0})

	newDispatcher(g).Dispatch(outbound.KeyEvent{Ch: '_'})

	assert.Equal(t, 1.5, g.Throttle().Speed())
}

// --- Unknown ---

func TestDispatcherUnknownRuneNoOp(t *testing.T) {
	g := fixture.Game(nil, game.Config{Speed: 2.0})
	d := newDispatcher(g)

	result := d.Dispatch(outbound.KeyEvent{Ch: 'z'})

	assert.Nil(t, result)
	assert.True(t, g.IsRunning())
	assert.Equal(t, 2.0, g.Throttle().Speed())
}
