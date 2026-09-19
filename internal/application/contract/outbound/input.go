package outbound

import "github.com/eirikur-ari/pidshooter/internal/application/input"

// InputEventSource is the outbound port for user input events.
type InputEventSource interface {
	// Events returns the channel of input events.
	// The channel is closed once no further events will be produced.
	Events() <-chan input.Event
}
