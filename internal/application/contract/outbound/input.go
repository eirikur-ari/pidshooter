package outbound

// InputEvent is implemented by all user input events.
type InputEvent interface{ isInputEvent() }

// ClickEvent represents a mouse click at terminal coordinates.
type ClickEvent struct{ X, Y int }

func (ClickEvent) isInputEvent() {
	// seals InputEvent to this package
}

// KeyEvent represents a key press.
type KeyEvent struct {
	Key KeyCode
	Ch  rune
}

func (KeyEvent) isInputEvent() {
	// seals InputEvent to this package
}

// KeyCode represents a named non-character key.
type KeyCode int

const (
	KeyNone KeyCode = iota
	KeyEscape
	KeyCtrlC
	KeyCtrlZ
)

// InputSource is the outbound port for user input events.
type InputSource interface {
	Events() <-chan InputEvent
}
