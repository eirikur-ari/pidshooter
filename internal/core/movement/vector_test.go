package movement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVector_Rounded_RoundsEachComponentHalfAwayFromZero(t *testing.T) {
	tests := []struct {
		name      string
		input     Vector
		expectedX int
		expectedY int
	}{
		{name: "below half rounds down", input: Vector{X: 2.4, Y: 5.4}, expectedX: 2, expectedY: 5},
		{name: "above half rounds up", input: Vector{X: 10.6, Y: 5.6}, expectedX: 11, expectedY: 6},
		{name: "exactly half rounds up", input: Vector{X: 0.5, Y: 1.5}, expectedX: 1, expectedY: 2},
		{name: "negative half rounds away from zero", input: Vector{X: -0.5, Y: -1.5}, expectedX: -1, expectedY: -2},
		{name: "whole numbers are unchanged", input: Vector{X: 3, Y: 7}, expectedX: 3, expectedY: 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			x, y := tt.input.Rounded()

			// Then
			assert.Equal(t, tt.expectedX, x)
			assert.Equal(t, tt.expectedY, y)
		})
	}
}
