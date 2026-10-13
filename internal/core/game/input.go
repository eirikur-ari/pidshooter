package game

// Input is the domain gateway for player intent-level input actions.
// The application layer maps raw input events to these calls.
type Input struct {
	s *Session
}

// NewInput returns an Input backed by the given session.
func NewInput(s *Session) *Input {
	return &Input{s: s}
}

// OnQuit cancels a pending kill confirmation if one is active, otherwise stops the session.
func (h *Input) OnQuit() {
	if h.s.confirm.Pending() {
		h.s.confirm.Cancel()
	} else {
		h.s.Stop()
	}
}

// OnYes accepts a pending kill confirmation and returns the target to kill.
func (h *Input) OnYes() *Target {
	return h.s.confirm.Accept()
}

// OnNo cancels a pending kill confirmation.
func (h *Input) OnNo() {
	h.s.confirm.Cancel()
}

// OnSpeedUp increases the session throttle.
func (h *Input) OnSpeedUp() {
	h.s.Throttle().Increase()
}

// OnSpeedDown decreases the session throttle.
func (h *Input) OnSpeedDown() {
	h.s.Throttle().Decrease()
}

// OnClickAt processes a click at the given coordinates and returns the target to kill if one is hit.
// In confirm mode, the first hit sets the pending confirmation rather than returning the target.
func (h *Input) OnClickAt(x, y int) *Target {
	if h.s.confirm.Pending() {
		return nil
	}

	target := h.s.roster.hitAt(x, y)

	if target == nil {
		return nil
	}

	return h.s.confirm.Request(target)
}
