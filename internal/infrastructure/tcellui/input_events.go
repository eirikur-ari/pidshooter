package tcellui

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
)

// Compile-time assertion that *inputEvents satisfies outbound.InputEventProvider.
var _ outbound.InputEventProvider = (*inputEvents)(nil)

// inputEvents implements outbound.InputEventProvider.
type inputEvents struct {
	poller *poller
}

// Events returns the channel of translated input events.
func (e *inputEvents) Events() <-chan input.EventDispatcher {
	return e.poller.events()
}
