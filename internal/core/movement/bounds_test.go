package movement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoundsBounceLeft(t *testing.T) {
	b := NewBounds(WindowSize{Width: 80, Height: 24}, ChromeSize{Top: 1, Bottom: 1})
	pos := Vector{X: -1.0, Y: 5.0}
	vel := Vector{X: -1.0, Y: 0}

	b.bounce(&pos, &vel, 5)

	assert.Equal(t, 0.0, pos.X)
	assert.Greater(t, vel.X, 0.0)
}

func TestBoundsBounceRight(t *testing.T) {
	b := NewBounds(WindowSize{Width: 80, Height: 24}, ChromeSize{Top: 1, Bottom: 1})
	pos := Vector{X: 76.0, Y: 5.0}
	vel := Vector{X: 1.0, Y: 0}

	b.bounce(&pos, &vel, 5)

	assert.LessOrEqual(t, pos.X, 75.0)
	assert.Less(t, vel.X, 0.0)
}

func TestBoundsBounceTop(t *testing.T) {
	b := NewBounds(WindowSize{Width: 80, Height: 24}, ChromeSize{Top: 1, Bottom: 1})
	pos := Vector{X: 5.0, Y: -1.0}
	vel := Vector{X: 0, Y: -1.0}

	b.bounce(&pos, &vel, 5)

	assert.Equal(t, 1.0, pos.Y)
	assert.Greater(t, vel.Y, 0.0)
}

func TestBoundsBounceBottomDoesNotUndoTopReservation(t *testing.T) {
	b := NewBounds(WindowSize{Width: 80, Height: 2}, ChromeSize{Top: 1, Bottom: 1})
	pos := Vector{X: 5.0, Y: -1.0}
	vel := Vector{X: 0, Y: -1.0}

	b.bounce(&pos, &vel, 5)

	assert.Equal(t, 1.0, pos.Y)
}

func TestBoundsBounceBottom(t *testing.T) {
	b := NewBounds(WindowSize{Width: 80, Height: 24}, ChromeSize{Top: 1, Bottom: 1})
	pos := Vector{X: 5.0, Y: 23.0}
	vel := Vector{X: 0, Y: 1.0}

	b.bounce(&pos, &vel, 5)

	assert.LessOrEqual(t, pos.Y, float64(24-2))
	assert.Less(t, vel.Y, 0.0)
}

func TestBoundsBounceTopRespectsGivenTopValue(t *testing.T) {
	b := NewBounds(WindowSize{Width: 80, Height: 24}, ChromeSize{Top: 3, Bottom: 1})
	pos := Vector{X: 5.0, Y: 2.0}
	vel := Vector{X: 0, Y: -1.0}

	b.bounce(&pos, &vel, 5)

	assert.Equal(t, 3.0, pos.Y)
	assert.Greater(t, vel.Y, 0.0)
}

func TestBoundsBounceBottomRespectsGivenBottomValue(t *testing.T) {
	b := NewBounds(WindowSize{Width: 80, Height: 24}, ChromeSize{Top: 1, Bottom: 4})
	pos := Vector{X: 5.0, Y: 20.0}
	vel := Vector{X: 0, Y: 1.0}

	b.bounce(&pos, &vel, 5)

	assert.Equal(t, 19.0, pos.Y) // height-1-bottom = 24-1-4
	assert.Less(t, vel.Y, 0.0)
}
