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
	tr := newTimer(60, time.Now)
	tr.Start()
	assert.False(t, tr.Expired())
}

func TestTimerExpiredTrueWhenLimitReached(t *testing.T) {
	clock := &fakeClock{t: time.Now()}
	tr := newTimer(1, clock.now)
	tr.Start()
	clock.advance(2 * time.Second)
	assert.True(t, tr.Expired())
}

func TestTimerExpiredFalseWhenUnlimited(t *testing.T) {
	tr := newTimer(0, time.Now)
	assert.False(t, tr.Expired())
}

func TestTimerRemainingWithinLimit(t *testing.T) {
	tr := newTimer(60, time.Now)
	tr.Start()
	r := tr.Remaining()
	assert.Greater(t, r, time.Duration(0))
	assert.LessOrEqual(t, r, 60*time.Second)
}

func TestTimerRemainingZeroWhenExpired(t *testing.T) {
	clock := &fakeClock{t: time.Now()}
	tr := newTimer(1, clock.now)
	tr.Start()
	clock.advance(2 * time.Second)
	assert.Equal(t, time.Duration(0), tr.Remaining())
}

func TestTimerRemainingZeroWhenUnlimited(t *testing.T) {
	tr := newTimer(0, time.Now)
	tr.Start()
	assert.Equal(t, time.Duration(0), tr.Remaining())
}

func TestTimerSecondsLeft(t *testing.T) {
	tr := newTimer(60, time.Now)
	tr.Start()
	assert.LessOrEqual(t, tr.SecondsLeft(), 60)
	assert.Greater(t, tr.SecondsLeft(), 0)
}

func TestTimerSecondsLeftRoundsUpPartialSecond(t *testing.T) {
	clock := &fakeClock{t: time.Now()}
	tr := newTimer(5, clock.now)
	tr.Start()
	clock.advance(4100 * time.Millisecond)

	assert.Equal(t, 1, tr.SecondsLeft(), "0.9s remaining should round up to 1s, not truncate to 0s")
}

func TestTimerSecondsLeftExactWholeSecondIsUnaffected(t *testing.T) {
	clock := &fakeClock{t: time.Now()}
	tr := newTimer(5, clock.now)
	tr.Start()
	clock.advance(2 * time.Second)

	assert.Equal(t, 3, tr.SecondsLeft())
}

func TestTimerSecondsLeftZeroWhenExpired(t *testing.T) {
	clock := &fakeClock{t: time.Now()}
	tr := newTimer(1, clock.now)
	tr.Start()
	clock.advance(2 * time.Second)

	assert.Equal(t, 0, tr.SecondsLeft())
}

func TestTimerLimitSeconds(t *testing.T) {
	tr30 := newTimer(30, time.Now)
	assert.Equal(t, 30, tr30.LimitSeconds())
	tr0 := newTimer(0, time.Now)
	assert.Equal(t, 0, tr0.LimitSeconds())
}

func TestTimerStartTimeZeroBeforeStart(t *testing.T) {
	tr := newTimer(60, time.Now)
	assert.True(t, tr.StartTime().IsZero())
}

func TestTimerStartTimeSetAfterStart(t *testing.T) {
	tr := newTimer(60, time.Now)
	before := time.Now()
	tr.Start()
	assert.False(t, tr.StartTime().Before(before))
}

// fakeClock is a controllable time source for deterministic timer tests.
type fakeClock struct {
	t time.Time
}

func (c *fakeClock) now() time.Time { return c.t }

func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }
