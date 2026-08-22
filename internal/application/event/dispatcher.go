package event

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

// Dispatcher translates raw input events into intent-level domain calls.
type Dispatcher struct {
	input *game.Input
}

// NewDispatcher returns a Dispatcher that dispatches to the given Input.
func NewDispatcher(input *game.Input) *Dispatcher {
	return &Dispatcher{input: input}
}

// Dispatch routes an input event to the appropriate domain call.
func (d *Dispatcher) Dispatch(event outbound.InputEvent) *game.Target {
	switch e := event.(type) {
	case outbound.ClickEvent:
		return d.handleClick(e)
	case outbound.KeyEvent:
		return d.handleKey(e)
	}
	return nil
}

func (d *Dispatcher) handleClick(event outbound.ClickEvent) *game.Target {
	return d.input.OnClickAt(event.X, event.Y)
}

func (d *Dispatcher) handleKey(event outbound.KeyEvent) *game.Target {
	if event.Key != outbound.KeyNone {
		return d.handleKeyCode(event.Key)
	}
	return d.handleRune(event.Ch)
}

func (d *Dispatcher) handleKeyCode(key outbound.KeyCode) *game.Target {
	switch key {
	case outbound.KeyEscape, outbound.KeyCtrlC, outbound.KeyCtrlZ:
		d.input.OnQuit()
	case outbound.KeyNone:
		// nothing to do
	}
	return nil
}

func (d *Dispatcher) handleRune(ch rune) *game.Target {
	switch ch {
	case 'q', 'Q':
		d.input.OnQuit()
	case 'y', 'Y':
		return d.input.OnYes()
	case 'n', 'N':
		d.input.OnNo()
	case '+', '=':
		d.input.OnSpeedUp()
	case '-', '_':
		d.input.OnSpeedDown()
	}
	return nil
}
