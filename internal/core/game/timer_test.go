package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimer_Expired_FalseBeforeLimit(t *testing.T) {
	tr := NewTimer(60)
	tr.Start()
	assert.False(t, tr.Expired())
}

func TestTimer_Expired_TrueWhenLimitReached(t *testing.T) {
	tr := NewTimer(1)
	tr.start = time.Now().Add(-2 * time.Second)
	assert.True(t, tr.Expired())
}

func TestTimer_Expired_FalseWhenUnlimited(t *testing.T) {
	tr := NewTimer(0)
	tr.start = time.Now().Add(-100 * time.Second)
	assert.False(t, tr.Expired())
}

func TestTimer_Remaining_WithinLimit(t *testing.T) {
	tr := NewTimer(60)
	tr.Start()
	r := tr.Remaining()
	assert.Greater(t, r, time.Duration(0))
	assert.LessOrEqual(t, r, 60*time.Second)
}

func TestTimer_Remaining_ZeroWhenExpired(t *testing.T) {
	tr := NewTimer(1)
	tr.start = time.Now().Add(-2 * time.Second)
	assert.Equal(t, time.Duration(0), tr.Remaining())
}

func TestTimer_Remaining_ZeroWhenUnlimited(t *testing.T) {
	tr := NewTimer(0)
	tr.Start()
	assert.Equal(t, time.Duration(0), tr.Remaining())
}

func TestTimer_SecondsLeft(t *testing.T) {
	tr := NewTimer(60)
	tr.Start()
	assert.LessOrEqual(t, tr.SecondsLeft(), 60)
	assert.Greater(t, tr.SecondsLeft(), 0)
}

func TestTimer_LimitSeconds(t *testing.T) {
	assert.Equal(t, 30, NewTimer(30).LimitSeconds())
	assert.Equal(t, 0, NewTimer(0).LimitSeconds())
}

func TestTimer_StartTime_ZeroBeforeStart(t *testing.T) {
	tr := NewTimer(60)
	assert.True(t, tr.StartTime().IsZero())
}

func TestTimer_StartTime_SetAfterStart(t *testing.T) {
	tr := NewTimer(60)
	before := time.Now()
	tr.Start()
	assert.False(t, tr.StartTime().Before(before))
}
