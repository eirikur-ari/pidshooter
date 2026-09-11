// Package score manages high-score ranking logic.
package score

import (
	"time"
)

// Entry represents a single high score record.
type Entry struct {
	Kills    int
	FreedMem int64
	Speed    float64
	// Time is the configured session time limit in seconds, distinct
	// from Duration (how long the session actually ran).
	Time     int
	Duration float64
	Date     time.Time
}

// beats ranks entries by kills, then speed, then duration, then freed memory.
func (e Entry) beats(other Entry) bool {
	if e.Kills != other.Kills {
		return e.Kills > other.Kills
	}
	if e.Speed != other.Speed {
		return e.Speed > other.Speed
	}
	if e.Duration != other.Duration {
		return e.Duration < other.Duration
	}
	return e.FreedMem > other.FreedMem
}
