package game

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

func TestKillTracker_recordKill_AccumulatesKillsAndFreedMemory(t *testing.T) {
	// Given
	tracker := newKillTracker(0)

	// When
	tracker.recordKill(1024)
	tracker.recordKill(2048)

	// Then
	assert.Equal(t, 2, tracker.score.kills)
	assert.Equal(t, int64(3072), tracker.score.freedMem)
}

func TestKillTracker_recordKill_RaisesHighScoreOnlyOnceKillsExceedIt(t *testing.T) {
	// Given
	tracker := newKillTracker(1)

	// When
	tracker.recordKill(0)

	// Then
	assert.Equal(t, 1, tracker.score.highScore, "kills equal to the high score do not raise it")

	// When
	tracker.recordKill(0)

	// Then
	assert.Equal(t, 2, tracker.score.highScore, "kills above the high score raise it")
}

func TestKillTracker_recordFailure_RecordsNamePIDAndCause(t *testing.T) {
	tests := newRecordFailureTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			tracker := newKillTracker(0)

			// When
			tracker.recordFailure(test.target, test.cause)

			// Then
			assert.Equal(t, []KillFailure{test.expected}, tracker.failure.failures)
		})
	}
}

func TestKillTracker_recordFailure_RecordsFailuresOfDistinctPIDs(t *testing.T) {
	// Given
	tracker := newKillTracker(0)

	// When
	tracker.recordFailure(newTargetFixtureFor(100, "one"), errors.New("boom"))
	tracker.recordFailure(newTargetFixtureFor(101, "two"), errors.New("another boom"))

	// Then
	assert.Len(t, tracker.failure.failures, 2)
}

func TestKillTracker_recordFailure_IgnoresRepeatFailureOfSamePID(t *testing.T) {
	// Given
	tracker := newKillTracker(0)
	target := newTargetFixture()
	first := errors.New("boom")

	// When
	tracker.recordFailure(target, first)
	tracker.recordFailure(target, errors.New("boom again"))

	// Then
	assert.Equal(t, []KillFailure{{Name: "target", PID: 100, Err: first}}, tracker.failure.failures)
}

func newRecordFailureTestCase() []struct {
	name     string
	target   *game.Target
	cause    error
	expected KillFailure
} {
	cause := errors.New("operation not permitted")

	return []struct {
		name     string
		target   *game.Target
		cause    error
		expected KillFailure
	}{
		{"without cause", newTargetFixtureFor(100, "target"), nil, KillFailure{Name: "target", PID: 100}},
		{"with cause", newTargetFixtureFor(200, "other"), cause, KillFailure{Name: "other", PID: 200, Err: cause}},
	}
}
