package game

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

// killTracker accumulates state for a single playthrough: score progress,
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
	messages []string
	pids     map[int]struct{}
}

// newKillTracker returns a killTracker seeded with the caller's
// persisted high score.
func newKillTracker(highScore int) *killTracker {
	return &killTracker{score: killScoreTracker{highScore: highScore}}
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
	if _, ok := t.failure.pids[target.Pid]; ok {
		return
	}
	if t.failure.pids == nil {
		t.failure.pids = make(map[int]struct{})
	}
	t.failure.pids[target.Pid] = struct{}{}
	t.failure.messages = append(t.failure.messages, fmt.Sprintf("could not kill %s (PID %d): %v", target.Name, target.Pid, err))
}
