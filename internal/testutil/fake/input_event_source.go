package fake

import "github.com/eirikur-ari/pidshooter/internal/application/input"

// InputEventSource is a test double for outbound.InputEventSource.
// Push events onto Ch to drive test behaviour.
type InputEventSource struct {
	Ch chan input.Event
}

// NewInputEventSource returns an InputEventSource with a buffered channel.
func NewInputEventSource() *InputEventSource {
	return &InputEventSource{Ch: make(chan input.Event, 100)}
}

func (e *InputEventSource) Events() <-chan input.Event { return e.Ch }
