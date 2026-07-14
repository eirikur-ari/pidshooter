package game

// Confirmation tracks a target pending user confirmation before a kill is issued.
// When confirm is false, Set passes targets through immediately.
type Confirmation struct {
	target  *Target
	confirm bool
}

// NewConfirmation returns a Confirmation configured for the given mode.
func NewConfirmation(confirm bool) Confirmation {
	return Confirmation{confirm: confirm}
}

// Pending reports whether a target is awaiting confirmation.
func (c *Confirmation) Pending() bool { return c.target != nil }

// Set registers t as pending confirmation (confirm mode) or returns it
// immediately for killing (passthrough mode).
func (c *Confirmation) Set(t *Target) *Target {
	if c.confirm {
		c.target = t
		return nil
	}
	return t
}

// Kill accepts the confirmation, clears the pending target, and returns it.
func (c *Confirmation) Kill() *Target { t := c.target; c.target = nil; return t }

// Clear cancels the pending confirmation without issuing a kill.
func (c *Confirmation) Clear() { c.target = nil }

// ViewState returns a ConfirmState for the pending target, or nil if none is pending.
func (c *Confirmation) ViewState() *ConfirmState {
	if c.target == nil {
		return nil
	}
	return &ConfirmState{PID: c.target.Pid, Name: c.target.Name}
}
