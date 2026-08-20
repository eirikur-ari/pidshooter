package event

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fixture"
)

func newDispatcher(s *game.Session) *Dispatcher {
	return NewDispatcher(game.NewInput(s))
}

// --- ClickEvent ---

func TestDispatcherClickReturnsHitTarget(t *testing.T) {
	s := fixture.Game(fixture.Processes(1), game.Config{Speed: 1.0})
	d := newDispatcher(s)
	tgt := s.Targets()[0]

	result := d.Dispatch(outbound.ClickEvent{X: int(math.Round(tgt.Position.X)), Y: int(math.Round(tgt.Position.Y))})

	assert.Equal(t, tgt, result)
}

func TestDispatcherClickMissReturnsNil(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 1.0})
	d := newDispatcher(s)

	result := d.Dispatch(outbound.ClickEvent{X: 0, Y: 0})

	assert.Nil(t, result)
}

// --- KeyCode ---

func TestDispatcherEscapeStopsGame(t *testing.T) {
	s := fixture.Game(nil, game.Config{})

	newDispatcher(s).Dispatch(outbound.KeyEvent{Key: outbound.KeyEscape})

	assert.False(t, s.IsRunning())
}

func TestDispatcherCtrlCStopsGame(t *testing.T) {
	s := fixture.Game(nil, game.Config{})

	newDispatcher(s).Dispatch(outbound.KeyEvent{Key: outbound.KeyCtrlC})

	assert.False(t, s.IsRunning())
}

func TestDispatcherCtrlZStopsGame(t *testing.T) {
	s := fixture.Game(nil, game.Config{})

	newDispatcher(s).Dispatch(outbound.KeyEvent{Key: outbound.KeyCtrlZ})

	assert.False(t, s.IsRunning())
}

// --- Rune: quit ---

func TestDispatcherQStopsGame(t *testing.T) {
	s := fixture.Game(nil, game.Config{})

	newDispatcher(s).Dispatch(outbound.KeyEvent{Ch: 'q'})

	assert.False(t, s.IsRunning())
}

func TestDispatcherQUppercaseStopsGame(t *testing.T) {
	s := fixture.Game(nil, game.Config{})

	newDispatcher(s).Dispatch(outbound.KeyEvent{Ch: 'Q'})

	assert.False(t, s.IsRunning())
}

// --- Rune: confirm yes ---

func TestDispatcherYReturnsConfirmedTarget(t *testing.T) {
	s := fixture.PendingConfirmGameSession(1)
	d := newDispatcher(s)
	tgt := s.Targets()[0]

	result := d.Dispatch(outbound.KeyEvent{Ch: 'y'})

	assert.Equal(t, tgt, result)
	assert.Nil(t, s.PendingConfirm())
}

func TestDispatcherYUppercaseAcceptsConfirmation(t *testing.T) {
	s := fixture.PendingConfirmGameSession(1)
	d := newDispatcher(s)
	tgt := s.Targets()[0]

	result := d.Dispatch(outbound.KeyEvent{Ch: 'Y'})

	assert.Equal(t, tgt, result)
	assert.Nil(t, s.PendingConfirm())
}

// --- Rune: confirm no ---

func TestDispatcherNCancelsConfirmation(t *testing.T) {
	s := fixture.PendingConfirmGameSession(1)
	d := newDispatcher(s)

	d.Dispatch(outbound.KeyEvent{Ch: 'n'})

	assert.Nil(t, s.PendingConfirm())
}

func TestDispatcherNUppercaseCancelsConfirmation(t *testing.T) {
	s := fixture.PendingConfirmGameSession(1)
	d := newDispatcher(s)

	d.Dispatch(outbound.KeyEvent{Ch: 'N'})

	assert.Nil(t, s.PendingConfirm())
}

// --- Rune: speed ---

func TestDispatcherPlusIncreasesSpeed(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 2.0})

	newDispatcher(s).Dispatch(outbound.KeyEvent{Ch: '+'})

	assert.Equal(t, 2.5, s.Throttle().Speed())
}

func TestDispatcherEqualsIncreasesSpeed(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 2.0})

	newDispatcher(s).Dispatch(outbound.KeyEvent{Ch: '='})

	assert.Equal(t, 2.5, s.Throttle().Speed())
}

func TestDispatcherMinusDecreasesSpeed(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 2.0})

	newDispatcher(s).Dispatch(outbound.KeyEvent{Ch: '-'})

	assert.Equal(t, 1.5, s.Throttle().Speed())
}

func TestDispatcherUnderscoreDecreasesSpeed(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 2.0})

	newDispatcher(s).Dispatch(outbound.KeyEvent{Ch: '_'})

	assert.Equal(t, 1.5, s.Throttle().Speed())
}

// --- Unknown ---

func TestDispatcherUnknownRuneNoOp(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 2.0})
	d := newDispatcher(s)

	result := d.Dispatch(outbound.KeyEvent{Ch: 'z'})

	assert.Nil(t, result)
	assert.True(t, s.IsRunning())
	assert.Equal(t, 2.0, s.Throttle().Speed())
}
