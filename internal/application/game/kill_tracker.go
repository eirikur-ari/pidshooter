package game

import (
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

// KillFailure records a target that was not killed, along with the error that caused the failure.
type KillFailure struct {
	Target string
	PID    int
	Err    error
}

// killTracker accumulates state of a single game session: score progress,
// and kill attempts that failed to kill their targets.
type killTracker struct {
	score   killScoreTracker
	failure killFailureTracker
}

type killScoreTracker struct {
	kills     int
	freedMem  int64
	highScore int
}

type killFailureTracker struct {
	failures []KillFailure
	pids     map[int]struct{}
}

// newKillTracker returns a killTracker seeded with the caller's
// persisted high score.
func newKillTracker(highScore int) *killTracker {
	return &killTracker{
		score:   killScoreTracker{highScore: highScore},
		failure: killFailureTracker{pids: make(map[int]struct{})},
	}
}

// recordKill records a kill, updates freed memory, and raises the high score when needed.
func (t *killTracker) recordKill(freedMemory int64) {
	t.score.kills++
	t.score.freedMem += freedMemory
	if t.score.kills > t.score.highScore {
		t.score.highScore = t.score.kills
	}
}

// recordFailure records a kill attempt that left the target alive, deduplicated by PID.
func (t *killTracker) recordFailure(target *game.Target, err error) {
	if _, ok := t.failure.pids[target.PID]; ok {
		return
	}
	t.failure.pids[target.PID] = struct{}{}
	t.failure.failures = append(t.failure.failures, KillFailure{Target: target.Name, PID: target.PID, Err: err})
}
