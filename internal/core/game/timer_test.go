package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateTimeLimitNegative(t *testing.T) {
	assert.Error(t, ValidateTimeLimit(-1))
}

func TestValidateTimeLimitZeroIsUnlimited(t *testing.T) {
	assert.NoError(t, ValidateTimeLimit(0))
}

func TestValidateTimeLimitPositive(t *testing.T) {
	assert.NoError(t, ValidateTimeLimit(30))
}

func TestValidateTimeLimitMaxBoundary(t *testing.T) {
	assert.NoError(t, ValidateTimeLimit(maxTimeLimitSeconds))
}

func TestValidateTimeLimitExceedsMax(t *testing.T) {
	assert.Error(t, ValidateTimeLimit(maxTimeLimitSeconds+1))
}

func TestTimerExpiredFalseBeforeLimit(t *testing.T) {
	tr := newTimer(60)
	tr.Start()
	assert.False(t, tr.Expired())
}

func TestTimerExpiredTrueWhenLimitReached(t *testing.T) {
	tr := newTimer(1)
	tr.start = time.Now().Add(-2 * time.Second)
	assert.True(t, tr.Expired())
}

func TestTimerExpiredFalseWhenUnlimited(t *testing.T) {
	tr := newTimer(0)
	tr.start = time.Now().Add(-100 * time.Second)
	assert.False(t, tr.Expired())
}

func TestTimerRemainingWithinLimit(t *testing.T) {
	tr := newTimer(60)
	tr.Start()
	r := tr.Remaining()
	assert.Greater(t, r, time.Duration(0))
	assert.LessOrEqual(t, r, 60*time.Second)
}

func TestTimerRemainingZeroWhenExpired(t *testing.T) {
	tr := newTimer(1)
	tr.start = time.Now().Add(-2 * time.Second)
	assert.Equal(t, time.Duration(0), tr.Remaining())
}

func TestTimerRemainingZeroWhenUnlimited(t *testing.T) {
	tr := newTimer(0)
	tr.Start()
	assert.Equal(t, time.Duration(0), tr.Remaining())
}

func TestTimerSecondsLeft(t *testing.T) {
	tr := newTimer(60)
	tr.Start()
	assert.LessOrEqual(t, tr.SecondsLeft(), 60)
	assert.Greater(t, tr.SecondsLeft(), 0)
}

func TestTimerSecondsLeftRoundsUpPartialSecond(t *testing.T) {
	tr := newTimer(5)
	start := time.Now()
	tr.start = start
	tr.now = func() time.Time { return start.Add(4100 * time.Millisecond) }

	assert.Equal(t, 1, tr.SecondsLeft(), "0.9s remaining should round up to 1s, not truncate to 0s")
}

func TestTimerSecondsLeftExactWholeSecondIsUnaffected(t *testing.T) {
	tr := newTimer(5)
	start := time.Now()
	tr.start = start
	tr.now = func() time.Time { return start.Add(2 * time.Second) }

	assert.Equal(t, 3, tr.SecondsLeft())
}

func TestTimerSecondsLeftZeroWhenExpired(t *testing.T) {
	tr := newTimer(1)
	tr.start = time.Now().Add(-2 * time.Second)

	assert.Equal(t, 0, tr.SecondsLeft())
}

func TestTimerLimitSeconds(t *testing.T) {
	tr30 := newTimer(30)
	assert.Equal(t, 30, tr30.LimitSeconds())
	tr0 := newTimer(0)
	assert.Equal(t, 0, tr0.LimitSeconds())
}

func TestTimerStartTimeZeroBeforeStart(t *testing.T) {
	tr := newTimer(60)
	assert.True(t, tr.StartTime().IsZero())
}

func TestTimerStartTimeSetAfterStart(t *testing.T) {
	tr := newTimer(60)
	before := time.Now()
	tr.Start()
	assert.False(t, tr.StartTime().Before(before))
}
