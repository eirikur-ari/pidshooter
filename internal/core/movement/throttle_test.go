package movement

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThrottle_LowestSpeed_KeepsFlooredValueAfterIncrease(t *testing.T) {
	// Given
	throttle := NewThrottle(0.8)

	// When
	throttle.Decrease()
	throttle.Increase()

	// Then
	assert.Equal(t, 1.0, throttle.Speed())
	assert.Equal(t, MinSpeed, throttle.LowestSpeed(), "the lowest speed is the floored value, not the unclamped 0.3")
}

func TestThrottle_Increase_AddsOneStepCappedAtMax(t *testing.T) {
	tests := newThrottleIncreaseTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			throttle := NewThrottle(test.start)

			// When
			throttle.Increase()

			// Then
			assert.Equal(t, test.expectedSpeed, throttle.Speed())
		})
	}
}

func TestThrottle_Decrease_SubtractsOneStepFlooredAtMin(t *testing.T) {
	tests := newThrottleDecreaseTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			throttle := NewThrottle(test.start)

			// When
			throttle.Decrease()

			// Then
			assert.Equal(t, test.expectedSpeed, throttle.Speed())
		})
	}
}

func TestValidateSpeed_RejectsOutOfRangeAndNonFiniteValues(t *testing.T) {
	tests := newInvalidSpeedTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			err := ValidateSpeed(test.value)

			// Then
			assert.Error(t, err)
		})
	}
}

func TestValidateSpeed_AcceptsValuesFromMinToMax(t *testing.T) {
	tests := newValidSpeedTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			err := ValidateSpeed(test.value)

			// Then
			assert.NoError(t, err)
		})
	}
}

func newThrottleIncreaseTestCase() []struct {
	name          string
	start         float64
	expectedSpeed float64
} {
	return []struct {
		name          string
		start         float64
		expectedSpeed float64
	}{
		{"adds one step", 2.0, 2.5},
		{"landing exactly on max", 4.5, MaxSpeed},
		{"capped when a step would pass max", 4.8, MaxSpeed},
		{"stays at max", MaxSpeed, MaxSpeed},
	}
}

func newThrottleDecreaseTestCase() []struct {
	name          string
	start         float64
	expectedSpeed float64
} {
	return []struct {
		name          string
		start         float64
		expectedSpeed float64
	}{
		{"subtracts one step", 2.0, 1.5},
		{"landing exactly on min", 1.0, MinSpeed},
		{"floored when a step would pass min", 0.8, MinSpeed},
		{"stays at min", MinSpeed, MinSpeed},
	}
}

func newInvalidSpeedTestCase() []struct {
	name  string
	value float64
} {
	return []struct {
		name  string
		value float64
	}{
		{"just below min", math.Nextafter(MinSpeed, 0)},
		{"just above max", math.Nextafter(MaxSpeed, math.Inf(1))},
		{"NaN", math.NaN()},
		{"positive Inf", math.Inf(1)},
		{"negative Inf", math.Inf(-1)},
	}
}

func newValidSpeedTestCase() []struct {
	name  string
	value float64
} {
	return []struct {
		name  string
		value float64
	}{
		{"min boundary", MinSpeed},
		{"max boundary", MaxSpeed},
		{"mid range", 2.0},
	}
}
