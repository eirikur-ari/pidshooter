package fake

import "github.com/eirikur-ari/pidshooter/internal/core/event"

// InputSource is a test double for outbound.InputSource.
// Push events onto Ch to drive test behaviour.
type InputSource struct {
	Ch chan event.InputEvent
}

// NewInputSource returns an InputSource with a buffered channel.
func NewInputSource() *InputSource {
	return &InputSource{Ch: make(chan event.InputEvent, 100)}
}

func (e *InputSource) Events() <-chan event.InputEvent { return e.Ch }
