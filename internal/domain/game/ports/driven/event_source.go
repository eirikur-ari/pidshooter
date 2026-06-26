package driven

// EventSource is the driven port for user input events.
type EventSource interface {
	Events() <-chan InputEvent
}

// InputEvent is a sealed interface for all input event types.
type InputEvent interface{ inputEvent() }

// ClickEvent represents a mouse click at terminal coordinates.
type ClickEvent struct{ X, Y int }

// KeyEvent represents a key press.
type KeyEvent struct {
	Key KeyCode
	Ch  rune
}

// ResizeEvent signals a terminal resize; dimensions are re-queried from Renderer.
type ResizeEvent struct{}

func (ClickEvent) inputEvent()  {}
func (KeyEvent) inputEvent()    {}
func (ResizeEvent) inputEvent() {}

// KeyCode represents a named non-character key.
type KeyCode int

const (
	KeyNone    KeyCode = iota
	KeyEscape
	KeyCtrlC
	KeyCtrlZ
)
