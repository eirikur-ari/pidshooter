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
	expired := tr.Expired()

	// Then
	assert.False(t, expired)
}

func TestTimer_Expired_IsTrueOnceLimitIsReached(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(60)

	// When
	clock.Advance(60 * time.Second)
	expired := tr.Expired()

	// Then
	assert.True(t, expired)
}

func TestTimer_Expired_ReturnsFalseWhenTimeIsUnlimited(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(0)

	// When
	clock.Advance(time.Hour)
	expired := tr.Expired()

	// Then
	assert.False(t, expired)
}

func TestTimer_Remaining_ReturnsTimeLeftWithinTimeLimit(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(60)

	// When
	clock.Advance(10 * time.Second)
	remaining := tr.Remaining()

	// Then
	assert.Equal(t, 50*time.Second, remaining)
}

func TestTimer_Remaining_IsZeroAfterTimeLimitPassed(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(1)

	// When
	clock.Advance(2 * time.Second)
	remaining := tr.Remaining()

	// Then
	assert.Equal(t, time.Duration(0), remaining)
}

func TestTimer_Remaining_IsZeroWhenTimeIsUnlimited(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(0)

	// When
	clock.Advance(time.Hour)
	remaining := tr.Remaining()

	// Then
	assert.Equal(t, time.Duration(0), remaining)
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
	secondsLeft := tr.SecondsLeft()

	// Then
	assert.Equal(t, 1, secondsLeft, "0.9s remaining should round up to 1s, not truncate to 0s")
}

func TestTimer_SecondsLeft_IsUnaffectedWhenRemainingIsWholeSeconds(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(5)

	// When
	clock.Advance(2 * time.Second)
	secondsLeft := tr.SecondsLeft()

	// Then
	assert.Equal(t, 3, secondsLeft)
}

func TestTimer_SecondsLeft_IsZeroAfterLimitPassed(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(1)

	// When
	clock.Advance(2 * time.Second)
	secondsLeft := tr.SecondsLeft()

	// Then
	assert.Equal(t, 0, secondsLeft)
}

func TestTimer_SecondsLeft_IsZeroWhenUnlimited(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(0)

	// When
	clock.Advance(time.Hour)
	secondsLeft := tr.SecondsLeft()

	// Then
	assert.Equal(t, 0, secondsLeft)
}

func TestTimer_LimitSeconds_ReturnsConfiguredLimit(t *testing.T) {
	tests := newLimitSecondsTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			tr := newTimer(test.limitSeconds, time.Now)

			// When
			result := tr.LimitSeconds()

			// Then
			assert.Equal(t, test.expectedLimitSeconds, result)
		})
	}
}

func TestTimer_StartTime_IsZeroBeforeStart(t *testing.T) {
	// Given
	tr := newTimer(60, time.Now)

	// When
	isZero := tr.StartTime().IsZero()

	// Then
	assert.True(t, isZero)
}

func TestTimer_StartTime_ReturnsInitialStartTime(t *testing.T) {
	// Given
	tr, clock := newStartedTimer(60)
	startedAt := clock.Now()

	// When
	clock.Advance(10 * time.Second)
	startTime := tr.StartTime()

	// Then
	assert.Equal(t, startedAt, startTime)
}

func TestValidateTimeLimit_RejectsNegativeAndAboveMax(t *testing.T) {
	tests := newInvalidTimeLimitTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			err := ValidateTimeLimit(test.limit)

			// Then
			assert.Error(t, err)
		})
	}
}

func TestValidateTimeLimit_AcceptsZeroAndUpToMax(t *testing.T) {
	tests := newValidTimeLimitTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			err := ValidateTimeLimit(test.limit)

			// Then
			assert.NoError(t, err)
		})
	}
}

func newLimitSecondsTestCase() []struct {
	name                 string
	limitSeconds         int
	expectedLimitSeconds int
} {
	return []struct {
		name                 string
		limitSeconds         int
		expectedLimitSeconds int
	}{
		{"limited", 30, 30},
		{"unlimited", 0, 0},
	}
}

func newInvalidTimeLimitTestCase() []struct {
	name  string
	limit int
} {
	return []struct {
		name  string
		limit int
	}{
		{"negative", -1},
		{"above max", maxTimeLimitSeconds + 1},
	}
}

func newValidTimeLimitTestCase() []struct {
	name  string
	limit int
} {
	return []struct {
		name  string
		limit int
	}{
		{"zero is unlimited", 0},
		{"positive", 30},
		{"max boundary", maxTimeLimitSeconds},
	}
}
