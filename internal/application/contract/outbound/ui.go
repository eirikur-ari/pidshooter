package outbound

import "github.com/eirikur-ari/pidshooter/internal/core/game"

// Renderer is the outbound port for terminal rendering.
type Renderer interface {
	Init() error
	Cleanup()
	Size() (width, height int)
	Render(frame game.Frame)
}

// InputEvent is the outbound port type for all user input events.
type InputEvent interface{}

// InputSource is the outbound port for user input events.
type InputSource interface {
	Events() <-chan InputEvent
}
