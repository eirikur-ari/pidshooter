package game

// ClickEvent represents a mouse click at terminal coordinates.
type ClickEvent struct{ X, Y int }

// KeyEvent represents a key press.
type KeyEvent struct {
	Key KeyCode
	Ch  rune
}

// ResizeEvent signals a terminal resize; dimensions are re-queried from the renderer.
type ResizeEvent struct{}

// KeyCode represents a named non-character key.
type KeyCode int

const (
	KeyNone  KeyCode = iota
	KeyEscape
	KeyCtrlC
	KeyCtrlZ
)
