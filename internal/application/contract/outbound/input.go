package outbound

// InputEvent is implemented by all user input events.
type InputEvent interface{ isInputEvent() }

// ClickEvent represents a mouse click at terminal coordinates.
type ClickEvent struct{ X, Y int }

func (ClickEvent) isInputEvent() {
	// seals InputEvent to this package
}

// QuitEvent requests that the game stop running.
type QuitEvent struct{}

func (QuitEvent) isInputEvent() {
	// seals InputEvent to this package
}

// ConfirmEvent answers a pending kill confirmation.
type ConfirmEvent struct{ Accept bool }

func (ConfirmEvent) isInputEvent() {
	// seals InputEvent to this package
}

// SpeedEvent requests a one-step change to the game speed.
type SpeedEvent struct {
	// Faster is true to speed up, false to slow down.
	Faster bool
}

func (SpeedEvent) isInputEvent() {
	// seals InputEvent to this package
}

// InputSource is the outbound port for user input events.
type InputSource interface {
	// Events returns the channel of input events.
	// The channel is closed once no further events will be produced.
	Events() <-chan InputEvent
}
