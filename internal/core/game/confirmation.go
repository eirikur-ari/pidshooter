package game

// Confirmation tracks a target pending user confirmation before a kill is issued.
type Confirmation struct {
	target *Target
}

// Pending reports whether a target is awaiting confirmation.
func (c *Confirmation) Pending() bool { return c.target != nil }

// Set registers a target as pending confirmation.
func (c *Confirmation) Set(t *Target) { c.target = t }

// Kill accepts the confirmation, clears the pending target, and returns it.
func (c *Confirmation) Kill() *Target { t := c.target; c.target = nil; return t }

// Clear cancels the pending confirmation without issuing a kill.
func (c *Confirmation) Clear() { c.target = nil }
