package game

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

func TestKillResult_applyTo_KillsTargetWhoseKillSucceeded(t *testing.T) {
	// Given
	target := newTargetFixture()
	tracker := newKillTracker(0)

	// When
	killResult{target: target}.applyTo(tracker)

	// Then
	assert.Equal(t, game.Killing, target.State)
	assert.Equal(t, 1, tracker.score.kills)
	assert.Equal(t, int64(4096), tracker.score.freedMem)
}

func TestKillResult_applyTo_ReapsTargetWhoseProcessWasAlreadyGone(t *testing.T) {
	// Given
	target := newTargetFixture()
	alreadyGone := apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, "PID 100 already exited", nil)
	tracker := newKillTracker(0)

	// When
	killResult{target: target, err: alreadyGone}.applyTo(tracker)

	// Then
	assert.Equal(t, game.Fleeing, target.State)
	assert.Equal(t, 0, tracker.score.kills)
	assert.Equal(t, []KillDud{{Name: "target", PID: 100}}, tracker.duds)
}

func TestKillResult_applyTo_IgnoresReapOfTargetThatIsNotAlive(t *testing.T) {
	// Given
	target := newTargetFixture()
	target.Kill()
	alreadyGone := apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, "PID 100 already exited", nil)
	tracker := newKillTracker(0)

	// When
	killResult{target: target, err: alreadyGone}.applyTo(tracker)

	// Then
	assert.Equal(t, game.Killing, target.State)
	assert.Empty(t, tracker.duds)
}

func TestKillResult_applyTo_RecordsFailureAndLeavesTargetAlive(t *testing.T) {
	// Given
	target := newTargetFixture()
	cause := errors.New("operation not permitted")
	tracker := newKillTracker(0)

	// When
	killResult{target: target, err: cause}.applyTo(tracker)

	// Then
	assert.Equal(t, game.Alive, target.State)
	assert.Equal(t, 0, tracker.score.kills)
	assert.Equal(t, []KillFailure{{Name: "target", PID: 100, Err: cause}}, tracker.failure.failures)
}

func TestKillResult_applyTo_RecordsRepeatedFailuresOfSamePIDOnce(t *testing.T) {
	// Given
	target := newTargetFixture()
	tracker := newKillTracker(0)

	// When
	killResult{target: target, err: errors.New("operation not permitted")}.applyTo(tracker)
	killResult{target: target, err: errors.New("operation not permitted")}.applyTo(tracker)

	// Then
	assert.Len(t, tracker.failure.failures, 1)
}
