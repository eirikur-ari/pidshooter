package input

import "github.com/eirikur-ari/pidshooter/internal/core/game"

// Dispatcher routes intent-level input events to domain calls.
type Dispatcher struct {
	input *game.Input
}

// NewDispatcher returns a Dispatcher that dispatches to the given Input.
func NewDispatcher(input *game.Input) *Dispatcher {
	return &Dispatcher{input: input}
}

// Dispatch routes an input event to the appropriate domain call.
func (d *Dispatcher) Dispatch(inputEvent Event) *game.Target {
	return inputEvent.apply(d.input)
}
