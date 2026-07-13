package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrameBounds_Bounce_Left(t *testing.T) {
	b := FrameBounds{Width: 80, Height: 24}
	pos := FrameVector{X: -1.0, Y: 5.0}
	vel := FrameVector{X: -1.0, Y: 0}

	b.Bounce(&pos, &vel, 5)

	assert.Equal(t, 0.0, pos.X)
	assert.Greater(t, vel.X, 0.0)
}

func TestFrameBounds_Bounce_Right(t *testing.T) {
	b := FrameBounds{Width: 80, Height: 24}
	pos := FrameVector{X: 76.0, Y: 5.0}
	vel := FrameVector{X: 1.0, Y: 0}

	b.Bounce(&pos, &vel, 5)

	assert.LessOrEqual(t, pos.X, 75.0)
	assert.Less(t, vel.X, 0.0)
}

func TestFrameBounds_Bounce_Top(t *testing.T) {
	b := FrameBounds{Width: 80, Height: 24}
	pos := FrameVector{X: 5.0, Y: -1.0}
	vel := FrameVector{X: 0, Y: -1.0}

	b.Bounce(&pos, &vel, 5)

	assert.Equal(t, 0.0, pos.Y)
	assert.Greater(t, vel.Y, 0.0)
}

func TestFrameBounds_Bounce_Bottom(t *testing.T) {
	b := FrameBounds{Width: 80, Height: 24}
	pos := FrameVector{X: 5.0, Y: 23.0}
	vel := FrameVector{X: 0, Y: 1.0}

	b.Bounce(&pos, &vel, 5)

	assert.LessOrEqual(t, pos.Y, float64(24-2))
	assert.Less(t, vel.Y, 0.0)
}
