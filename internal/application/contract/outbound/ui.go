package outbound

import (
	"github.com/eirikur-ari/pidshooter/internal/core/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

// Renderer is the outbound port for terminal rendering.
type Renderer interface {
	Init() error
	Cleanup()
	Size() (width, height int)
	Render(frame game.FrameState)
}

// InputSource is the outbound port for user input events.
type InputSource interface {
	Events() <-chan event.InputEvent
}
