package movement

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVector_Rounded_RoundsEachComponentToNearestCellWithHalvesAwayFromZero(t *testing.T) {
	tests := []struct {
		name  string
		input Vector
		wantX int
		wantY int
	}{
		{name: "below half rounds down", input: Vector{X: 2.4, Y: 5.4}, wantX: 2, wantY: 5},
		{name: "above half rounds up", input: Vector{X: 10.6, Y: 5.6}, wantX: 11, wantY: 6},
		{name: "exactly half rounds up", input: Vector{X: 0.5, Y: 1.5}, wantX: 1, wantY: 2},
		{name: "negative half rounds away from zero", input: Vector{X: -0.5, Y: -1.5}, wantX: -1, wantY: -2},
		{name: "whole numbers are unchanged", input: Vector{X: 3, Y: 7}, wantX: 3, wantY: 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			x, y := tt.input.Rounded()

			// Then
			assert.Equal(t, tt.wantX, x)
			assert.Equal(t, tt.wantY, y)
		})
	}
}

func TestMotion_NewMotion_SpawnsWithinBounds(t *testing.T) {
	// Given
	bounds := newBoundsFixture()

	for range 200 {
		// When
		motion := NewMotion(bounds, 5)

		// Then
		assert.GreaterOrEqual(t, motion.Position.X, 1.0)
		assert.LessOrEqual(t, motion.Position.X, 74.0, "the rightmost spawn column leaves room for the width: 80-5-1")
		assert.GreaterOrEqual(t, motion.Position.Y, 1.0)
		assert.LessOrEqual(t, motion.Position.Y, 22.0, "the lowest spawn row is height-1-bottom = 24-1-1")
	}
}

func TestMotion_NewMotion_RespectsGivenChromeRows(t *testing.T) {
	// Given
	bounds := NewBounds(windowSizeFixture(), ChromeSize{Top: 5, Bottom: 6})

	for range 200 {
		// When
		motion := NewMotion(bounds, 5)

		// Then
		assert.GreaterOrEqual(t, motion.Position.Y, 5.0)
		assert.LessOrEqual(t, motion.Position.Y, 17.0, "height-1-bottom = 24-1-6")
	}
}

func TestMotion_NewMotion_SpawnsWithVelocityMagnitudeInRange(t *testing.T) {
	// Given
	bounds := newBoundsFixture()

	for range 200 {
		// When
		motion := NewMotion(bounds, 5)

		// Then
		assert.GreaterOrEqual(t, math.Abs(motion.Velocity.X), 0.5)
		assert.Less(t, math.Abs(motion.Velocity.X), 1.0)
		assert.GreaterOrEqual(t, math.Abs(motion.Velocity.Y), 0.25)
		assert.Less(t, math.Abs(motion.Velocity.Y), 0.5)
	}
}

func TestMotion_Move_AdvancesPositionByVelocityScaledBySpeed(t *testing.T) {
	// Given
	motion := Motion{
		Position: Vector{X: 10, Y: 5},
		Velocity: Vector{X: 1.0, Y: 1.0},
	}

	// When
	motion.Move(newBoundsFixture(), 3.0, 1)

	// Then
	assert.Equal(t, 13.0, motion.Position.X)
	assert.Equal(t, 8.0, motion.Position.Y)
}

func TestMotion_Move_BouncesAtRightWallByGivenWidth(t *testing.T) {
	// Given
	motion := Motion{
		Position: Vector{X: 75, Y: 5},
		Velocity: Vector{X: 2.0, Y: 0},
	}

	// When
	motion.Move(newBoundsFixture(), 1.0, 5)

	// Then
	assert.Equal(t, 75.0, motion.Position.X, "the right wall is the window width minus the given width: 80-5")
	assert.Equal(t, -2.0, motion.Velocity.X)
}
