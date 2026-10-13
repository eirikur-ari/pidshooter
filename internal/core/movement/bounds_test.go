package movement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBounds_Update_ReplacesWindowAndKeepsChromeWithoutMutatingOriginal(t *testing.T) {
	// Given
	chrome := ChromeSize{Top: 3, Bottom: 4}
	newWindow := WindowSize{Width: 40, Height: 10}
	original := NewBounds(windowSizeFixture(), chrome)
	untouched := original
	expected := NewBounds(newWindow, chrome)

	// When
	updated := original.Update(newWindow)

	// Then
	assert.Equal(t, expected, updated)
	assert.Equal(t, untouched, original, "Update must return a copy, not mutate the receiver")
}

func TestBounds_bounce_LeavesPositionAndVelocityUnchangedWhenWithinBounds(t *testing.T) {
	// Given
	bounds := newBoundsFixture()
	position := Vector{X: 40, Y: 10}
	velocity := Vector{X: 1.5, Y: -0.5}
	expectedPosition := Vector{X: 40, Y: 10}
	expectedVelocity := Vector{X: 1.5, Y: -0.5}

	// When
	bounds.bounce(&position, &velocity, 5)

	// Then
	assert.Equal(t, expectedPosition, position)
	assert.Equal(t, expectedVelocity, velocity)
}

func TestBounds_bounce_ReflectsAtLeftWall(t *testing.T) {
	// Given
	bounds := newBoundsFixture()
	position := Vector{X: -1, Y: 5}
	velocity := Vector{X: -1, Y: 0}

	// When
	bounds.bounce(&position, &velocity, 5)

	// Then
	assert.Equal(t, 0.0, position.X)
	assert.Equal(t, 1.0, velocity.X)
}

func TestBounds_bounce_ReflectsAtRightWall(t *testing.T) {
	// Given
	bounds := newBoundsFixture()
	position := Vector{X: 76, Y: 5}
	velocity := Vector{X: 1, Y: 0}

	// When
	bounds.bounce(&position, &velocity, 5)

	// Then
	assert.Equal(t, 75.0, position.X, "the right wall is the window width minus the moving width")
	assert.Equal(t, -1.0, velocity.X)
}

func TestBounds_bounce_ReflectsAtTopWall(t *testing.T) {
	// Given
	bounds := newBoundsFixture()
	position := Vector{X: 5, Y: -1}
	velocity := Vector{X: 0, Y: -1}

	// When
	bounds.bounce(&position, &velocity, 5)

	// Then
	assert.Equal(t, 1.0, position.Y)
	assert.Equal(t, 1.0, velocity.Y)
}

func TestBounds_bounce_ReflectsAtBottomWall(t *testing.T) {
	// Given
	bounds := newBoundsFixture()
	position := Vector{X: 5, Y: 23}
	velocity := Vector{X: 0, Y: 1}

	// When
	bounds.bounce(&position, &velocity, 5)

	// Then
	assert.Equal(t, 22.0, position.Y, "the bottom wall is the last row above the bottom chrome")
	assert.Equal(t, -1.0, velocity.Y)
}

func TestBounds_bounce_ClampsPositionButKeepsVelocityWhenAlreadyHeadingAwayFromWall(t *testing.T) {
	tests := newClampedBounceTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			bounds := newBoundsFixture()
			position, velocity := test.position, test.velocity

			// When
			bounds.bounce(&position, &velocity, 5)

			// Then
			assert.Equal(t, test.expectedPosition, position)
			assert.Equal(t, test.velocity, velocity, "velocity already pointing away from the wall must not be flipped back")
		})
	}
}

func TestBounds_bounce_ReflectsBothAxesWhenHittingACorner(t *testing.T) {
	// Given
	bounds := newBoundsFixture()
	position := Vector{X: -1, Y: -1}
	velocity := Vector{X: -1, Y: -1}
	expectedPosition := Vector{X: 0, Y: 1}
	expectedVelocity := Vector{X: 1, Y: 1}

	// When
	bounds.bounce(&position, &velocity, 5)

	// Then
	assert.Equal(t, expectedPosition, position)
	assert.Equal(t, expectedVelocity, velocity)
}

func TestBounds_bounce_PinsPositionToLeftEdgeWhenWidthExceedsWindow(t *testing.T) {
	// Given
	bounds := newBoundsFixture()
	position := Vector{X: 5, Y: 5}
	velocity := Vector{X: 0, Y: 0}

	// When
	bounds.bounce(&position, &velocity, 100)

	// Then
	assert.Equal(t, 0.0, position.X, "a width wider than the window leaves no room, so the right wall must not go negative")
}

func TestBounds_bounce_DoesNotLetBottomWallUndoTopReservationOnShortWindow(t *testing.T) {
	// Given
	bounds := NewBounds(WindowSize{Width: 80, Height: 2}, chromeSizeFixture())
	position := Vector{X: 5, Y: -1}
	velocity := Vector{X: 0, Y: -1}

	// When
	bounds.bounce(&position, &velocity, 5)

	// Then
	assert.Equal(t, 1.0, position.Y)
	assert.Equal(t, 1.0, velocity.Y)
}

func TestBounds_bounce_RespectsGivenTopChromeRows(t *testing.T) {
	// Given
	bounds := NewBounds(windowSizeFixture(), ChromeSize{Top: 3, Bottom: 1})
	position := Vector{X: 5, Y: 2}
	velocity := Vector{X: 0, Y: -1}

	// When
	bounds.bounce(&position, &velocity, 5)

	// Then
	assert.Equal(t, 3.0, position.Y)
	assert.Equal(t, 1.0, velocity.Y)
}

func TestBounds_bounce_RespectsGivenBottomChromeRows(t *testing.T) {
	// Given
	bounds := NewBounds(windowSizeFixture(), ChromeSize{Top: 1, Bottom: 4})
	position := Vector{X: 5, Y: 20}
	velocity := Vector{X: 0, Y: 1}

	// When
	bounds.bounce(&position, &velocity, 5)

	// Then
	assert.Equal(t, 19.0, position.Y, "height-1-bottom = 24-1-4")
	assert.Equal(t, -1.0, velocity.Y)
}

func newClampedBounceTestCase() []struct {
	name             string
	position         Vector
	velocity         Vector
	expectedPosition Vector
} {
	return []struct {
		name             string
		position         Vector
		velocity         Vector
		expectedPosition Vector
	}{
		{"left wall", Vector{X: -1, Y: 5}, Vector{X: 1, Y: 0}, Vector{X: 0, Y: 5}},
		{"right wall", Vector{X: 76, Y: 5}, Vector{X: -1, Y: 0}, Vector{X: 75, Y: 5}},
		{"top wall", Vector{X: 5, Y: -1}, Vector{X: 0, Y: 1}, Vector{X: 5, Y: 1}},
		{"bottom wall", Vector{X: 5, Y: 23}, Vector{X: 0, Y: -1}, Vector{X: 5, Y: 22}},
	}
}
