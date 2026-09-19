package fake

import "github.com/eirikur-ari/pidshooter/internal/application/input"

// InputEventProvider is a test double for outbound.InputEventProvider.
// Push events onto Ch to drive test behaviour.
type InputEventProvider struct {
	Ch chan input.EventDispatcher
}

// NewInputEventProvider returns an InputEventProvider with a buffered channel.
func NewInputEventProvider() *InputEventProvider {
	return &InputEventProvider{Ch: make(chan input.EventDispatcher, 100)}
}

func (e *InputEventProvider) Events() <-chan input.EventDispatcher { return e.Ch }
