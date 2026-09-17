package input

import "github.com/eirikur-ari/pidshooter/internal/core/game"

// Event is implemented by all user input events.
type Event interface {
	// apply routes the event to the appropriate call on input.
	apply(input *game.Input) *game.Target
}

// ClickEvent represents a mouse click at terminal coordinates.
type ClickEvent struct{ X, Y int }

func (e ClickEvent) apply(input *game.Input) *game.Target {
	return input.OnClickAt(e.X, e.Y)
}

// QuitEvent requests that the game stop running.
type QuitEvent struct{}

func (QuitEvent) apply(input *game.Input) *game.Target {
	input.OnQuit()
	return nil
}

// ConfirmEvent answers a pending kill confirmation.
type ConfirmEvent struct{ Accept bool }

func (e ConfirmEvent) apply(input *game.Input) *game.Target {
	if e.Accept {
		return input.OnYes()
	}
	input.OnNo()
	return nil
}

// SpeedEvent requests a one-step change to the game speed.
type SpeedEvent struct {
	// Faster is true to speed up, false to slow down.
	Faster bool
}

func (e SpeedEvent) apply(input *game.Input) *game.Target {
	if e.Faster {
		input.OnSpeedUp()
	} else {
		input.OnSpeedDown()
	}
	return nil
}
