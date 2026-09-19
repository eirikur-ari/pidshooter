package outbound

import "github.com/eirikur-ari/pidshooter/internal/application/input"

// InputEventProvider is the outbound port for user input events.
type InputEventProvider interface {
	// Events returns the channel of input events.
	// The channel is closed once no further events will be produced.
	Events() <-chan input.EventDispatcher
}
