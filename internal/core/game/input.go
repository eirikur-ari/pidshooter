package game

// Input is the domain gateway for player intent-level input actions.
// The application layer maps raw input events to these calls.
type Input struct {
	g *Game
}

// NewInput returns an Input backed by the given game.
func NewInput(g *Game) *Input {
	return &Input{g: g}
}

// OnQuit cancels a pending kill confirmation if one is active, otherwise stops the game.
func (h *Input) OnQuit() {
	if h.g.Confirm().Pending() {
		h.g.Confirm().Cancel()
	} else {
		h.g.Stop()
	}
}

// OnYes accepts a pending kill confirmation and returns the target to kill.
func (h *Input) OnYes() *Target {
	return h.g.Confirm().Accept()
}

// OnNo cancels a pending kill confirmation.
func (h *Input) OnNo() {
	h.g.Confirm().Cancel()
}

// OnSpeedUp increases the game throttle.
func (h *Input) OnSpeedUp() {
	h.g.Throttle().Increase()
}

// OnSpeedDown decreases the game throttle.
func (h *Input) OnSpeedDown() {
	h.g.Throttle().Decrease()
}

// OnClickAt processes a click at the given coordinates and returns the target to kill if one is hit.
// In confirm mode, the first hit sets the pending confirmation rather than returning the target.
func (h *Input) OnClickAt(x, y int) *Target {
	if h.g.Confirm().Pending() {
		return nil
	}
	for _, t := range h.g.Targets() {
		if t.isHitAt(x, y) {
			return h.g.Confirm().Request(t)
		}
	}
	return nil
}
