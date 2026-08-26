package event

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

// Dispatcher routes intent-level input events to domain calls.
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
		return d.input.OnClickAt(e.X, e.Y)
	case outbound.QuitEvent:
		d.input.OnQuit()
	case outbound.ConfirmEvent:
		if e.Accept {
			return d.input.OnYes()
		}
		d.input.OnNo()
	case outbound.SpeedEvent:
		if e.Faster {
			d.input.OnSpeedUp()
		} else {
			d.input.OnSpeedDown()
		}
	}
	return nil
}
