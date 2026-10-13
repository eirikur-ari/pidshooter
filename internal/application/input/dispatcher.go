package input

import "github.com/eirikur-ari/pidshooter/internal/core/game"

// Dispatcher applies input events to the Input it was created with.
type Dispatcher struct {
	input *game.Input
}

// NewDispatcher returns a Dispatcher that applies events to input.
func NewDispatcher(input *game.Input) *Dispatcher {
	return &Dispatcher{input: input}
}

// Dispatch applies the event to the game and returns the target it affected, or nil if none.
func (d *Dispatcher) Dispatch(inputEvent Event) *game.Target {
	return inputEvent.dispatch(d.input)
}
