package game

import "time"

// Timer tracks how much time remains in a timed game session.
// A zero limit means no time limit; Expired always returns false.
type Timer struct {
	limit time.Duration
	start time.Time
	now   func() time.Time
}

// NewTimer creates a Timer with the given limit in seconds. Zero means unlimited.
func NewTimer(limitSeconds int) Timer {
	return Timer{
		limit: time.Duration(limitSeconds) * time.Second,
		now:   time.Now,
	}
}

// Start records the session start time.
func (t *Timer) Start() {
	if t.now == nil {
		t.now = time.Now
	}
	t.start = t.now()
}

// Expired reports whether the time limit has been reached.
// Always false when no limit is set.
func (t *Timer) Expired() bool {
	return t.limit > 0 && t.Remaining() == 0
}

// Remaining returns the time left in the session.
// Returns 0 when expired or when no limit is set.
func (t *Timer) Remaining() time.Duration {
	if t.limit <= 0 {
		return 0
	}
	now := t.now
	if now == nil {
		now = time.Now
	}
	r := t.limit - now().Sub(t.start)
	if r < 0 {
		return 0
	}
	return r
}

// SecondsLeft returns the whole seconds remaining, for display.
// Returns 0 when there is no limit.
func (t *Timer) SecondsLeft() int {
	return int(t.Remaining().Seconds())
}

// LimitSeconds returns the configured time limit in seconds. Zero means unlimited.
func (t *Timer) LimitSeconds() int {
	return int(t.limit.Seconds())
}

// StartTime returns when the timer was started.
func (t *Timer) StartTime() time.Time {
	return t.start
}
