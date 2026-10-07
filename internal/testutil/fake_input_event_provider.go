package testutil

import "github.com/eirikur-ari/pidshooter/internal/application/input"

// FakeInputEventProvider is a test double for outbound.InputEventProvider.
// Push events onto Ch to drive test behavior.
type FakeInputEventProvider struct {
	Ch chan input.Event
}

// NewFakeInputEventProvider returns a FakeInputEventProvider with a buffered channel.
func NewFakeInputEventProvider() *FakeInputEventProvider {
	return &FakeInputEventProvider{Ch: make(chan input.Event, 100)}
}

func (e *FakeInputEventProvider) Events() <-chan input.Event { return e.Ch }
