package event

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

// Dispatcher translates raw input events into intent-level domain calls.
type Dispatcher struct {
	h *game.Input
}

// NewDispatcher returns a Dispatcher that dispatches to the given Input.
func NewDispatcher(h *game.Input) *Dispatcher {
	return &Dispatcher{h: h}
}

// Dispatch routes an input event to the appropriate domain call.
func (d *Dispatcher) Dispatch(ev outbound.InputEvent) *game.Target {
	switch ev := ev.(type) {
	case outbound.ClickEvent:
		return d.handleClick(ev)
	case outbound.KeyEvent:
		return d.handleKey(ev)
	}
	return nil
}

func (d *Dispatcher) handleClick(ev outbound.ClickEvent) *game.Target {
	return d.h.OnClickAt(ev.X, ev.Y)
}

func (d *Dispatcher) handleKey(ev outbound.KeyEvent) *game.Target {
	if ev.Key != outbound.KeyNone {
		return d.handleKeyCode(ev.Key)
	}
	return d.handleRune(ev.Ch)
}

func (d *Dispatcher) handleKeyCode(key outbound.KeyCode) *game.Target {
	switch key {
	case outbound.KeyEscape, outbound.KeyCtrlC, outbound.KeyCtrlZ:
		d.h.OnQuit()
	case outbound.KeyNone:
		// nothing to do
	}
	return nil
}

func (d *Dispatcher) handleRune(ch rune) *game.Target {
	switch ch {
	case 'q', 'Q':
		d.h.OnQuit()
	case 'y', 'Y':
		return d.h.OnYes()
	case 'n', 'N':
		d.h.OnNo()
	case '+', '=':
		d.h.OnSpeedUp()
	case '-', '_':
		d.h.OnSpeedDown()
	}
	return nil
}
