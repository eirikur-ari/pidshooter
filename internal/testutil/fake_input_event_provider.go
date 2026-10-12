package testutil

import "github.com/eirikur-ari/pidshooter/internal/application/input"

type FakeInputEventProvider struct {
	Ch chan input.Event
}

func NewFakeInputEventProvider() *FakeInputEventProvider {
	return &FakeInputEventProvider{Ch: make(chan input.Event, 100)}
}

func (e *FakeInputEventProvider) Events() <-chan input.Event { return e.Ch }
