package game

import "time"

// FakeClock is a controllable source of the current time.
type FakeClock struct {
	T time.Time
}

// Now returns the clock's current time.
func (c *FakeClock) Now() time.Time { return c.T }

// Advance moves the clock's current time forward by d.
func (c *FakeClock) Advance(d time.Duration) { c.T = c.T.Add(d) }
