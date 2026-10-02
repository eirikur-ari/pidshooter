package game

// confirmation tracks a target pending user confirmation.
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

// Request registers target as pending confirmation and returns nil, or, when
// confirmation is not required, returns target immediately.
func (c *confirmation) Request(target *Target) *Target {
	if c.confirm {
		c.target = target
		return nil
	}
	return target
}

// Accept returns the pending target, if any, and clears the pending confirmation.
func (c *confirmation) Accept() *Target {
	target := c.target
	c.target = nil
	return target
}

// Cancel clears the pending confirmation without accepting it.
func (c *confirmation) Cancel() { c.target = nil }
