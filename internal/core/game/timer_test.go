package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimer_Expired_IsFalseWhileAnyTimeRemains(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(60)

	// When
	clock.Advance(60*time.Second - time.Millisecond)

	// Then
	assert.False(t, tr.Expired())
}

func TestTimer_Expired_IsTrueOnceLimitIsReached(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(60)

	// When
	clock.Advance(60 * time.Second)

	// Then
	assert.True(t, tr.Expired())
}

func TestTimer_Expired_ReturnsFalseWhenTimeIsUnlimited(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(0)

	// When
	clock.Advance(time.Hour)

	// Then
	assert.False(t, tr.Expired())
}

func TestTimer_Remaining_ReturnsTimeLeftWithinTimeLimit(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(60)

	// When
	clock.Advance(10 * time.Second)

	// Then
	assert.Equal(t, 50*time.Second, tr.Remaining())
}

func TestTimer_Remaining_IsZeroAfterTimeLimitPassed(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(1)

	// When
	clock.Advance(2 * time.Second)

	// Then
	assert.Equal(t, time.Duration(0), tr.Remaining())
}

func TestTimer_Remaining_IsZeroWhenTimeIsUnlimited(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(0)

	// When
	clock.Advance(time.Hour)

	// Then
	assert.Equal(t, time.Duration(0), tr.Remaining())
}

func TestTimer_SecondsLeft_ReturnsWholeLimitAtStart(t *testing.T) {
	// Given
	tr, _ := newStartedTimer(60)

	// When
	result := tr.SecondsLeft()

	// Then
	assert.Equal(t, 60, result)
}

func TestTimer_SecondsLeft_RoundsUpPartialSecond(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(5)

	// When
	clock.Advance(4100 * time.Millisecond)

	// When
	assert.Equal(t, 1, tr.SecondsLeft(), "0.9s remaining should round up to 1s, not truncate to 0s")
}

func TestTimer_SecondsLeft_IsUnaffectedWhenRemainingIsWholeSeconds(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(5)

	// When
	clock.Advance(2 * time.Second)

	// Then
	assert.Equal(t, 3, tr.SecondsLeft())
}

func TestTimer_SecondsLeft_IsZeroAfterLimitPassed(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(1)

	// When
	clock.Advance(2 * time.Second)

	// Then
	assert.Equal(t, 0, tr.SecondsLeft())
}

func TestTimer_SecondsLeft_IsZeroWhenUnlimited(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(0)

	// When
	clock.Advance(time.Hour)

	// Then
	assert.Equal(t, 0, tr.SecondsLeft())
}

func TestTimer_LimitSeconds_ReturnsConfiguredLimit(t *testing.T) {
	for name, limit := range map[string]int{
		"limited":   30,
		"unlimited": 0,
	} {
		t.Run(name, func(t *testing.T) {
			// Given
			tr := newTimer(limit, time.Now)

			// When
			result := tr.LimitSeconds()

			// Then
			assert.Equal(t, limit, result)
		})
	}
}

func TestTimer_StartTime_IsZeroBeforeStart(t *testing.T) {
	// Given
	tr := newTimer(60, time.Now)

	// Then
	assert.True(t, tr.StartTime().IsZero())
}

func TestTimer_StartTime_ReturnsInitialStartTime(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(60)
	startedAt := clock.Now()

	// When
	clock.Advance(10 * time.Second)

	// Then
	assert.Equal(t, startedAt, tr.StartTime())
}

func TestValidateTimeLimit_RejectsNegativeAndAboveMax(t *testing.T) {
	for name, limit := range map[string]int{
		"negative":  -1,
		"above max": maxTimeLimitSeconds + 1,
	} {
		t.Run(name, func(t *testing.T) {
			// When
			err := ValidateTimeLimit(limit)

			// Then
			assert.Error(t, err)
		})
	}
}

func TestValidateTimeLimit_AcceptsZeroAndUpToMax(t *testing.T) {
	for name, limit := range map[string]int{
		"zero is unlimited": 0,
		"positive":          30,
		"max boundary":      maxTimeLimitSeconds,
	} {
		t.Run(name, func(t *testing.T) {
			// When
			err := ValidateTimeLimit(limit)

			// Then
			assert.NoError(t, err)
		})
	}
}

func newStartedTimer(limitSeconds int) (*timer, *FakeClock) {
	clock := &FakeClock{T: time.Date(2026, time.October, 4, 12, 0, 0, 0, time.UTC)}
	tr := newTimer(limitSeconds, clock.Now)
	tr.Start()
	return &tr, clock
}
