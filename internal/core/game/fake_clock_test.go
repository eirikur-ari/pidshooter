package game

import "time"

type FakeClock struct {
	T time.Time
}

func (c *FakeClock) Now() time.Time { return c.T }

func (c *FakeClock) Advance(d time.Duration) { c.T = c.T.Add(d) }
