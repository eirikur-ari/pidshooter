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

func TestKillTrackerCreatesNewKillTracker(t *testing.T) {
	tr := newKillTracker(5)
	assert.Equal(t, 5, tr.score.highScore)
	assert.Equal(t, 0, tr.score.kills)
	assert.Equal(t, int64(0), tr.score.freedMem)
	assert.NotNil(t, tr.failure.pids)
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

func TestKillTrackerRecordFailureAddsFailedKillWithNoCause(t *testing.T) {
	tr := newKillTracker(0)

	tr.recordFailure(newTrackerTestTarget(100, "target"), nil)

	require.Len(t, tr.failure.failures, 1)
	failure := tr.failure.failures[0]
	assert.Equal(t, "target", failure.Target)
	assert.Equal(t, 100, failure.PID)
}

func TestKillTrackerRecordFailureAddsFailedKillWithCause(t *testing.T) {
	tr := newKillTracker(0)
	cause := errors.New("operation not permitted")

	tr.recordFailure(newTrackerTestTarget(100, "target"), cause)

	require.Len(t, tr.failure.failures, 1)
	failure := tr.failure.failures[0]
	assert.Equal(t, cause, failure.Err)
}

func TestKillTrackerRecordFailureAccumulatesDistinctPIDs(t *testing.T) {
	tr := newKillTracker(0)

	tr.recordFailure(newTrackerTestTarget(100, "one"), errors.New("boom"))
	tr.recordFailure(newTrackerTestTarget(101, "two"), errors.New("another boom"))

	assert.Len(t, tr.failure.failures, 2)
}

func TestKillTrackerRecordFailureDeduplicatesSamePID(t *testing.T) {
	tr := newKillTracker(0)
	target := newTrackerTestTarget(100, "target")
	err := errors.New("boom")

	tr.recordFailure(target, err)
	tr.recordFailure(target, errors.New("boom again"))

	assert.Len(t, tr.failure.failures, 1, "a repeat failure for the same PID should not add a second entry")
	assert.Equal(t, err, tr.failure.failures[0].Err, "the second call should be a no-op, not an update")
}

func newTrackerTestTarget(pid int, name string) *game.Target {
	return game.NewTarget(process.NewInfo(pid, name, 4096, 0), movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
}
