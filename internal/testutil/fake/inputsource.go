package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// InputSource is a test double for outbound.InputSource.
// Push events onto Ch to drive test behaviour.
type InputSource struct {
	Ch chan outbound.InputEvent
}

// NewInputSource returns an InputSource with a buffered channel.
func NewInputSource() *InputSource {
	return &InputSource{Ch: make(chan outbound.InputEvent, 100)}
}

func (e *InputSource) Events() <-chan outbound.InputEvent { return e.Ch }
