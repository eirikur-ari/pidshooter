package game

import (
	"fmt"
	"math"
	"time"
)

// maxTimeLimitSeconds is the highest non-zero limit ValidateTimeLimit accepts.
const maxTimeLimitSeconds = 5 * 60 // 300 seconds = 5 minutes

// timer tracks how much time remains in a timed game session.
// A zero limit means no time limit.
type timer struct {
	limit time.Duration
	start time.Time
	now   func() time.Time
}

// newTimer creates a timer with the given limit in seconds and clock. Zero
// limit means unlimited.
func newTimer(limitSeconds int, now func() time.Time) timer {
	return timer{limit: secondsToDuration(limitSeconds), now: now}
}

// Start records the session start time.
func (t *timer) Start() {
	t.start = t.now()
}

// Expired reports whether the time limit has been reached.
// Always false when no limit is set.
func (t *timer) Expired() bool {
	return t.limit > 0 && t.Remaining() == 0
}

// Remaining returns the time left in the session.
// Returns 0 when expired or when no limit is set.
func (t *timer) Remaining() time.Duration {
	if t.limit <= 0 {
		return 0
	}
	duration := t.limit - t.now().Sub(t.start)
	if duration < 0 {
		return 0
	}
	return duration
}

// SecondsLeft returns the whole seconds remaining, rounded up so a session
// isn't reported as having 0 seconds left before it has actually expired.
// Returns 0 when there is no limit.
func (t *timer) SecondsLeft() int {
	return int(math.Ceil(t.Remaining().Seconds()))
}

// LimitSeconds returns the configured time limit in seconds. Zero means unlimited.
func (t *timer) LimitSeconds() int {
	return int(t.limit.Seconds())
}

// StartTime returns when the timer was started.
func (t *timer) StartTime() time.Time {
	return t.start
}

// ValidateTimeLimit reports an error if limit is negative or exceeds
// maxTimeLimitSeconds. Zero means unlimited.
func ValidateTimeLimit(limit int) error {
	if limit < 0 || limit > maxTimeLimitSeconds {
		return fmt.Errorf("time must be 0 (no limit) or between 1 and %d seconds, got: %d seconds", maxTimeLimitSeconds, limit)
	}
	return nil
}

// secondsToDuration converts a count of seconds to a time.Duration.
func secondsToDuration(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}
