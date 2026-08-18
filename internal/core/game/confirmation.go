package game

// Confirmation tracks a target pending user confirmation before a kill is issued.
// When confirm is false, Request passes targets through immediately.
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

// Request registers t as pending confirmation (confirm mode) or returns it
// immediately for killing (passthrough mode).
func (c *Confirmation) Request(t *Target) *Target {
	if c.confirm {
		c.target = t
		return nil
	}
	return t
}

// Accept accepts the pending confirmation, clears it, and returns the target.
func (c *Confirmation) Accept() *Target { t := c.target; c.target = nil; return t }

// Cancel cancels the pending confirmation without issuing a kill.
func (c *Confirmation) Cancel() { c.target = nil }
