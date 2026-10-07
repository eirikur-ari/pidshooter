package input

import "github.com/eirikur-ari/pidshooter/internal/core/game"

// Event is implemented by all user input events.
type Event interface {
	// dispatch applies the event to input and returns the target it affected, or nil if none.
	dispatch(input *game.Input) *game.Target
}

// ClickEvent represents a click at a position.
type ClickEvent struct {
	// X is the horizontal position of the click.
	X int
	// Y is the vertical position of the click.
	Y int
}

func (e ClickEvent) dispatch(input *game.Input) *game.Target {
	return input.OnClickAt(e.X, e.Y)
}

// QuitEvent represents a request to stop the game.
type QuitEvent struct{}

func (QuitEvent) dispatch(input *game.Input) *game.Target {
	input.OnQuit()
	return nil
}

// ConfirmEvent represents an answer to a confirmation request.
type ConfirmEvent struct {
	// Accept is true to accept, false to decline.
	Accept bool
}

func (e ConfirmEvent) dispatch(input *game.Input) *game.Target {
	if e.Accept {
		return input.OnYes()
	}
	input.OnNo()
	return nil
}

// SpeedEvent represents a request to change the game speed by one step.
type SpeedEvent struct {
	// Faster is true to speed up, false to slow down.
	Faster bool
}

func (e SpeedEvent) dispatch(input *game.Input) *game.Target {
	if e.Faster {
		input.OnSpeedUp()
	} else {
		input.OnSpeedDown()
	}
	return nil
}
