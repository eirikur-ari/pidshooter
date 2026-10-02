package movement

import (
	"math"
	"math/rand"
)

// Vector is a 2D positional or velocity vector.
type Vector struct {
	X, Y float64
}

// Rounded returns the vector's coordinates as the nearest integer cell.
func (v Vector) Rounded() (x, y int) {
	return int(math.Round(v.X)), int(math.Round(v.Y))
}

// Motion holds the position and velocity of a moving entity and encapsulates its physics.
type Motion struct {
	Position Vector
	Velocity Vector
}

// NewMotion returns a Motion with a random position within bounds and a random velocity.
// tagWidth is the rune width of the entity's display tag, used to constrain spawn position.
func NewMotion(bounds Bounds, tagWidth int) Motion {
	return Motion{
		Position: newRandomPosition(bounds, tagWidth),
		Velocity: newRandomVelocity(),
	}
}

// Move advances Position by Velocity scaled by speed, then bounces off the frame walls.
// tagWidth is the current rune width of the entity's display tag.
func (m *Motion) Move(bounds Bounds, speed, tagWidth float64) {
	m.Position.X += m.Velocity.X * speed
	m.Position.Y += m.Velocity.Y * speed
	bounds.bounce(&m.Position, &m.Velocity, tagWidth)
}

func newRandomPosition(bounds Bounds, tagWidth int) Vector {
	x := max(bounds.window.Width-tagWidth-1, 1)
	y := max(bounds.window.Height-bounds.chrome.Top-bounds.chrome.Bottom, 1)
	return Vector{X: float64(rand.Intn(x) + 1), Y: float64(rand.Intn(y) + bounds.chrome.Top)}
}

func newRandomVelocity() Vector {
	x := rand.Float64()*0.5 + 0.5
	if rand.Intn(2) == 0 {
		x = -x
	}
	y := rand.Float64()*0.25 + 0.25
	if rand.Intn(2) == 0 {
		y = -y
	}
	return Vector{X: x, Y: y}
}
