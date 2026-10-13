package movement

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMotion_SpawnsWithinBounds(t *testing.T) {
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

func TestNewMotion_RespectsGivenChromeRows(t *testing.T) {
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

func TestNewMotion_SpawnsWithVelocityMagnitudeInRange(t *testing.T) {
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
