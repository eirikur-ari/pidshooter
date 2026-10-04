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
	tests := []struct {
		name  string
		start float64
		speed float64
	}{
		{name: "adds one step", start: 2.0, speed: 2.5},
		{name: "landing exactly on max", start: 4.5, speed: MaxSpeed},
		{name: "capped when a step would pass max", start: 4.8, speed: MaxSpeed},
		{name: "stays at max", start: MaxSpeed, speed: MaxSpeed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			throttle := NewThrottle(tt.start)

			// When
			throttle.Increase()

			// Then
			assert.Equal(t, tt.speed, throttle.Speed())
		})
	}
}

func TestThrottle_Decrease_SubtractsOneStepFlooredAtMin(t *testing.T) {
	tests := []struct {
		name  string
		start float64
		speed float64
	}{
		{name: "subtracts one step", start: 2.0, speed: 1.5},
		{name: "landing exactly on min", start: 1.0, speed: MinSpeed},
		{name: "floored when a step would pass min", start: 0.8, speed: MinSpeed},
		{name: "stays at min", start: MinSpeed, speed: MinSpeed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			throttle := NewThrottle(tt.start)

			// When
			throttle.Decrease()

			// Then
			assert.Equal(t, tt.speed, throttle.Speed())
		})
	}
}

func TestValidateSpeed_RejectsOutOfRangeAndNonFiniteValues(t *testing.T) {
	tests := map[string]float64{
		"just below min": math.Nextafter(MinSpeed, 0),
		"just above max": math.Nextafter(MaxSpeed, math.Inf(1)),
		"NaN":            math.NaN(),
		"positive Inf":   math.Inf(1),
		"negative Inf":   math.Inf(-1),
	}
	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			// When
			err := ValidateSpeed(value)

			// Then
			assert.Error(t, err)
		})
	}
}

func TestValidateSpeed_AcceptsValuesFromMinToMax(t *testing.T) {
	tests := map[string]float64{
		"min boundary": MinSpeed,
		"max boundary": MaxSpeed,
		"mid range":    2.0,
	}
	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			// When
			err := ValidateSpeed(value)

			// Then
			assert.NoError(t, err)
		})
	}
}
