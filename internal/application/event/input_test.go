package event

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func newDispatcher(fh *fake.InputHandler) *Dispatcher {
	return NewDispatcher(fh)
}

// --- ClickEvent ---

func TestDispatcher_Click_CallsOnClickAt(t *testing.T) {
	fh := &fake.InputHandler{}
	d := newDispatcher(fh)

	d.Dispatch(outbound.ClickEvent{X: 5, Y: 10})

	assert.Equal(t, 1, fh.ClickCalled)
	assert.Equal(t, 5, fh.ClickX)
	assert.Equal(t, 10, fh.ClickY)
}

func TestDispatcher_Click_ReturnsTarget(t *testing.T) {
	tgt := game.NewTarget(process.NewInfo(1, "a", 0), movement.NewBounds(80, 24))
	fh := &fake.InputHandler{ClickResult: tgt}
	d := newDispatcher(fh)

	result := d.Dispatch(outbound.ClickEvent{X: 5, Y: 10})

	assert.Equal(t, tgt, result)
}

// --- KeyCode ---

func TestDispatcher_Escape_CallsOnQuit(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Key: outbound.KeyEscape})
	assert.Equal(t, 1, fh.QuitCalled)
}

func TestDispatcher_CtrlC_CallsOnQuit(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Key: outbound.KeyCtrlC})
	assert.Equal(t, 1, fh.QuitCalled)
}

func TestDispatcher_CtrlZ_CallsOnQuit(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Key: outbound.KeyCtrlZ})
	assert.Equal(t, 1, fh.QuitCalled)
}

// --- Rune: quit ---

func TestDispatcher_Q_CallsOnQuit(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Ch: 'q'})
	assert.Equal(t, 1, fh.QuitCalled)
}

func TestDispatcher_QUppercase_CallsOnQuit(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Ch: 'Q'})
	assert.Equal(t, 1, fh.QuitCalled)
}

// --- Rune: confirm yes ---

func TestDispatcher_Y_CallsOnYes(t *testing.T) {
	tgt := game.NewTarget(process.NewInfo(1, "a", 0), movement.NewBounds(80, 24))
	fh := &fake.InputHandler{YesResult: tgt}
	d := newDispatcher(fh)

	result := d.Dispatch(outbound.KeyEvent{Ch: 'y'})

	assert.Equal(t, 1, fh.YesCalled)
	assert.Equal(t, tgt, result)
}

func TestDispatcher_YUppercase_CallsOnYes(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Ch: 'Y'})
	assert.Equal(t, 1, fh.YesCalled)
}

// --- Rune: confirm no ---

func TestDispatcher_N_CallsOnNo(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Ch: 'n'})
	assert.Equal(t, 1, fh.NoCalled)
}

func TestDispatcher_NUppercase_CallsOnNo(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Ch: 'N'})
	assert.Equal(t, 1, fh.NoCalled)
}

// --- Rune: speed ---

func TestDispatcher_Plus_CallsOnSpeedUp(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Ch: '+'})
	assert.Equal(t, 1, fh.SpeedUpCalled)
}

func TestDispatcher_Equals_CallsOnSpeedUp(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Ch: '='})
	assert.Equal(t, 1, fh.SpeedUpCalled)
}

func TestDispatcher_Minus_CallsOnSpeedDown(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Ch: '-'})
	assert.Equal(t, 1, fh.SpeedDownCalled)
}

func TestDispatcher_Underscore_CallsOnSpeedDown(t *testing.T) {
	fh := &fake.InputHandler{}
	newDispatcher(fh).Dispatch(outbound.KeyEvent{Ch: '_'})
	assert.Equal(t, 1, fh.SpeedDownCalled)
}

// --- Unknown ---

func TestDispatcher_UnknownRune_NoOp(t *testing.T) {
	fh := &fake.InputHandler{}
	d := newDispatcher(fh)

	result := d.Dispatch(outbound.KeyEvent{Ch: 'z'})

	assert.Nil(t, result)
	assert.Equal(t, 0, fh.QuitCalled+fh.YesCalled+fh.NoCalled+fh.SpeedUpCalled+fh.SpeedDownCalled)
}
