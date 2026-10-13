package game

import (
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

// killTracker holds the progress of a single game session.
type killTracker struct {
	score   killScoreTracker
	failure killFailureTracker
	duds    []killDud
}

type killScoreTracker struct {
	kills     int
	freedMem  int64
	highScore int
}

type killFailureTracker struct {
	failures []killFailure
	pids     map[int]struct{}
}

// newKillTracker returns a killTracker whose high score starts at highScore.
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
	if _, ok := t.failure.pids[target.Info.PID]; ok {
		return
	}
	t.failure.pids[target.Info.PID] = struct{}{}
	t.failure.failures = append(t.failure.failures, killFailure{Name: target.Info.Name, PID: target.Info.PID, Err: err})
}

// recordDud records a target whose backing process was already gone before
// a kill could land on it.
func (t *killTracker) recordDud(target *game.Target) {
	t.duds = append(t.duds, killDud{Name: target.Info.Name, PID: target.Info.PID})
}

func (t *killTracker) kills() int { return t.score.kills }

func (t *killTracker) freedMem() int64 { return t.score.freedMem }

func (t *killTracker) highScore() int { return t.score.highScore }

// errors returns an error for each recorded failure, followed by one for each recorded dud.
func (t *killTracker) errors() []error {
	var errs []error
	for _, failure := range t.failure.failures {
		errs = append(errs, failure.toError())
	}
	for _, dud := range t.duds {
		errs = append(errs, dud.toError())
	}
	return errs
}
