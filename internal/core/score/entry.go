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
	Time     int
	Duration float64
	Date     time.Time
}

func (e Entry) beats(other Entry) bool {
	if e.Kills != other.Kills {
		return e.Kills > other.Kills
	}
	return e.FreedMem > other.FreedMem
}
