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
	if h.g.confirm.Pending() {
		h.g.confirm.Cancel()
	} else {
		h.g.Stop()
	}
}

// OnYes accepts a pending kill confirmation and returns the target to kill.
func (h *Input) OnYes() *Target {
	return h.g.confirm.Accept()
}

// OnNo cancels a pending kill confirmation.
func (h *Input) OnNo() {
	h.g.confirm.Cancel()
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
	if h.g.confirm.Pending() {
		return nil
	}
	t := h.g.roster.hitAt(x, y)
	if t == nil {
		return nil
	}
	return h.g.confirm.Request(t)
}
