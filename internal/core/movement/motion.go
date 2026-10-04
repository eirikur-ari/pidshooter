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

// Motion holds the position and velocity of a moving object.
type Motion struct {
	Position Vector
	Velocity Vector
}

// NewMotion returns a Motion with a random position within bounds and a random velocity.
// width is the horizontal extent of what is moving, used to constrain spawn position.
func NewMotion(bounds Bounds, width int) Motion {
	return Motion{
		Position: newRandomPosition(bounds, width),
		Velocity: newRandomVelocity(),
	}
}

// Move advances Position by Velocity scaled by speed, then bounces off the bounds' walls.
// width is the horizontal extent of what is moving.
func (m *Motion) Move(bounds Bounds, speed float64, width int) {
	m.Position.X += m.Velocity.X * speed
	m.Position.Y += m.Velocity.Y * speed
	bounds.bounce(&m.Position, &m.Velocity, width)
}

func newRandomPosition(bounds Bounds, width int) Vector {
	x := max(bounds.window.Width-width-1, 1)
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
