package game

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func newTrackerTestTarget(pid int, name string) *game.Target {
	return game.NewTarget(process.NewInfo(pid, name, 4096), movement.NewBounds(80, 24))
}

func TestNewKillTrackerSeedsHighScore(t *testing.T) {
	tr := newKillTracker(5)
	assert.Equal(t, 5, tr.score.highScore)
	assert.Equal(t, 0, tr.score.kills)
	assert.Equal(t, int64(0), tr.score.freedMem)
}

func TestKillTrackerRecordKillAccumulates(t *testing.T) {
	tr := newKillTracker(0)

	tr.recordKill(1024)
	tr.recordKill(2048)

	assert.Equal(t, 2, tr.score.kills)
	assert.Equal(t, int64(3072), tr.score.freedMem)
}

func TestKillTrackerRecordKillRaisesHighScore(t *testing.T) {
	tr := newKillTracker(1)

	tr.recordKill(0)
	assert.Equal(t, 1, tr.score.highScore, "kills should not exceed the seeded high score yet")

	tr.recordKill(0)
	assert.Equal(t, 2, tr.score.highScore, "high score should rise once kills exceed it")
}

func TestKillTrackerRecordFailureAddsMessage(t *testing.T) {
	tr := newKillTracker(0)

	tr.recordFailure(newTrackerTestTarget(100, "target"), errors.New("operation not permitted"))

	require.Len(t, tr.failure.messages, 1)
	assert.Equal(t, "could not kill target (PID 100): operation not permitted", tr.failure.messages[0])
}

func TestKillTrackerRecordFailureAccumulatesDistinctPIDs(t *testing.T) {
	tr := newKillTracker(0)

	tr.recordFailure(newTrackerTestTarget(100, "one"), errors.New("boom"))
	tr.recordFailure(newTrackerTestTarget(101, "two"), errors.New("boom"))

	assert.Len(t, tr.failure.messages, 2)
}

func TestKillTrackerRecordFailureDeduplicatesSamePID(t *testing.T) {
	tr := newKillTracker(0)
	target := newTrackerTestTarget(100, "target")

	tr.recordFailure(target, errors.New("boom"))
	tr.recordFailure(target, errors.New("boom again"))

	assert.Len(t, tr.failure.messages, 1, "a repeat failure for the same PID should not add a second entry")
}
