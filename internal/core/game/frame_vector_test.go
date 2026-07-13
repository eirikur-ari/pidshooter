package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrameVector_Apply(t *testing.T) {
	vel := FrameVector{X: 1.0, Y: 0.5}
	pos := FrameVector{X: 10.0, Y: 5.0}

	vel.Apply(&pos, 2.0)

	assert.Equal(t, 12.0, pos.X)
	assert.Equal(t, 6.0, pos.Y)
}

func TestFrameVector_Apply_NegativeVelocity(t *testing.T) {
	vel := FrameVector{X: -1.0, Y: -0.5}
	pos := FrameVector{X: 10.0, Y: 5.0}

	vel.Apply(&pos, 1.0)

	assert.Equal(t, 9.0, pos.X)
	assert.Equal(t, 4.5, pos.Y)
}
