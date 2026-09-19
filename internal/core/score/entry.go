// Package score manages high-score ranking logic.
package score

import (
	"time"
)

// Entry represents a single high score record.
type Entry struct {
	Kills int
	// Duds is the number of targets whose backing process was already gone
	// before a kill could land on it.
	Duds     int
	FreedMem int64
	Speed    float64
	// Time is the configured session time limit in seconds, distinct
	// from Duration (how long the session actually ran).
	Time     int
	Duration float64
	Date     time.Time
}

// beats ranks entries by kills, then duds (more duds wins a kills tie), then
// speed, then duration, then freed memory.
func (e Entry) beats(other Entry) bool {
	if e.Kills != other.Kills {
		return e.Kills > other.Kills
	}
	if e.Duds != other.Duds {
		return e.Duds > other.Duds
	}
	if e.Speed != other.Speed {
		return e.Speed > other.Speed
	}
	if e.Duration != other.Duration {
		return e.Duration < other.Duration
	}
	return e.FreedMem > other.FreedMem
}
