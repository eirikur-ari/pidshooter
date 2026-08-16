package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimerExpiredFalseBeforeLimit(t *testing.T) {
	tr := NewTimer(60)
	tr.Start()
	assert.False(t, tr.Expired())
}

func TestTimerExpiredTrueWhenLimitReached(t *testing.T) {
	tr := NewTimer(1)
	tr.start = time.Now().Add(-2 * time.Second)
	assert.True(t, tr.Expired())
}

func TestTimerExpiredFalseWhenUnlimited(t *testing.T) {
	tr := NewTimer(0)
	tr.start = time.Now().Add(-100 * time.Second)
	assert.False(t, tr.Expired())
}

func TestTimerRemainingWithinLimit(t *testing.T) {
	tr := NewTimer(60)
	tr.Start()
	r := tr.Remaining()
	assert.Greater(t, r, time.Duration(0))
	assert.LessOrEqual(t, r, 60*time.Second)
}

func TestTimerRemainingZeroWhenExpired(t *testing.T) {
	tr := NewTimer(1)
	tr.start = time.Now().Add(-2 * time.Second)
	assert.Equal(t, time.Duration(0), tr.Remaining())
}

func TestTimerRemainingZeroWhenUnlimited(t *testing.T) {
	tr := NewTimer(0)
	tr.Start()
	assert.Equal(t, time.Duration(0), tr.Remaining())
}

func TestTimerSecondsLeft(t *testing.T) {
	tr := NewTimer(60)
	tr.Start()
	assert.LessOrEqual(t, tr.SecondsLeft(), 60)
	assert.Greater(t, tr.SecondsLeft(), 0)
}

func TestTimerLimitSeconds(t *testing.T) {
	tr30 := NewTimer(30)
	assert.Equal(t, 30, tr30.LimitSeconds())
	tr0 := NewTimer(0)
	assert.Equal(t, 0, tr0.LimitSeconds())
}

func TestTimerStartTimeZeroBeforeStart(t *testing.T) {
	tr := NewTimer(60)
	assert.True(t, tr.StartTime().IsZero())
}

func TestTimerStartTimeSetAfterStart(t *testing.T) {
	tr := NewTimer(60)
	before := time.Now()
	tr.Start()
	assert.False(t, tr.StartTime().Before(before))
}
