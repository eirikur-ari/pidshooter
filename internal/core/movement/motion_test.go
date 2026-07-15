package movement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMotion_Move_AdvancesPosition(t *testing.T) {
	m := Motion{
		Position: Vector{X: 10, Y: 5},
		Velocity: Vector{X: 1.0, Y: 0.5},
	}

	m.Move(NewBounds(80, 24), 2.0, 5)

	assert.Equal(t, 12.0, m.Position.X)
	assert.Equal(t, 6.0, m.Position.Y)
}

func TestMotion_Move_BouncesAtLeftWall(t *testing.T) {
	m := Motion{
		Position: Vector{X: 0, Y: 5},
		Velocity: Vector{X: -1.0, Y: 0},
	}

	m.Move(NewBounds(80, 24), 1.0, 5)

	assert.GreaterOrEqual(t, m.Position.X, 0.0)
	assert.Greater(t, m.Velocity.X, 0.0)
}

func TestMotion_Move_BouncesAtRightWall(t *testing.T) {
	m := Motion{
		Position: Vector{X: 75, Y: 5},
		Velocity: Vector{X: 2.0, Y: 0},
	}

	m.Move(NewBounds(80, 24), 1.0, 5)

	assert.LessOrEqual(t, m.Position.X, 75.0)
	assert.Less(t, m.Velocity.X, 0.0)
}

func TestMotion_Move_SpeedScalesVelocity(t *testing.T) {
	m := Motion{
		Position: Vector{X: 10, Y: 5},
		Velocity: Vector{X: 1.0, Y: 1.0},
	}

	m.Move(NewBounds(80, 24), 3.0, 1)

	assert.Equal(t, 13.0, m.Position.X)
	assert.Equal(t, 8.0, m.Position.Y)
}
