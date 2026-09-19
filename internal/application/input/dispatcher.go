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

// Dispatch delegates the event to its own dispatch logic.
func (d *Dispatcher) Dispatch(inputEvent EventDispatcher) *game.Target {
	return inputEvent.dispatch(d.input)
}
