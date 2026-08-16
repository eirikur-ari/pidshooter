package handler

import "github.com/eirikur-ari/pidshooter/internal/core/game"

// Handler is the domain gateway for player intent-level input actions.
// The application layer maps raw input events to these calls.
type Handler struct {
	g *game.Game
}

// NewHandler returns a Handler backed by the given game.
func NewHandler(g *game.Game) *Handler {
	return &Handler{g: g}
}

// OnQuit cancels a pending kill confirmation if one is active, otherwise stops the game.
func (h *Handler) OnQuit() {
	if h.g.Confirm().Pending() {
		h.g.Confirm().Cancel()
	} else {
		h.g.Stop()
	}
}

// OnYes accepts a pending kill confirmation and returns the target to kill.
func (h *Handler) OnYes() *game.Target {
	return h.g.Confirm().Accept()
}

// OnNo cancels a pending kill confirmation.
func (h *Handler) OnNo() {
	h.g.Confirm().Cancel()
}

// OnSpeedUp increases the game throttle.
func (h *Handler) OnSpeedUp() {
	h.g.Throttle().Increase()
}

// OnSpeedDown decreases the game throttle.
func (h *Handler) OnSpeedDown() {
	h.g.Throttle().Decrease()
}

// OnClickAt processes a click at the given coordinates and returns the target to kill if one is hit.
// In confirm mode, the first hit sets the pending confirmation rather than returning the target.
func (h *Handler) OnClickAt(x, y int) *game.Target {
	if h.g.Confirm().Pending() {
		return nil
	}
	for _, t := range h.g.Targets() {
		if t.IsHitAt(x, y) {
			return h.g.Confirm().Request(t)
		}
	}
	return nil
}
