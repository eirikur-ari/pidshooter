package game

// confirmation tracks a target pending user confirmation before a kill is issued.
// When confirm is false, Request passes targets through immediately.
type confirmation struct {
	target  *Target
	confirm bool
}

// newConfirmation returns a confirmation configured for the given mode.
func newConfirmation(confirm bool) confirmation {
	return confirmation{confirm: confirm}
}

// Pending reports whether a target is awaiting confirmation.
func (c *confirmation) Pending() bool { return c.target != nil }

// Request registers t as pending confirmation (confirm mode) or returns it
// immediately for killing (passthrough mode).
func (c *confirmation) Request(t *Target) *Target {
	if c.confirm {
		c.target = t
		return nil
	}
	return t
}

// Accept accepts the pending confirmation, clears it, and returns the target.
func (c *confirmation) Accept() *Target { t := c.target; c.target = nil; return t }

// Cancel cancels the pending confirmation without issuing a kill.
func (c *confirmation) Cancel() { c.target = nil }
