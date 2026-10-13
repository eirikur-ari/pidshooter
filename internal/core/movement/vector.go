package movement

import "math"

// Vector is a 2D positional or velocity vector.
type Vector struct {
	X, Y float64
}

// Rounded returns the vector's coordinates as the nearest integer cell.
func (v Vector) Rounded() (x, y int) {
	return int(math.Round(v.X)), int(math.Round(v.Y))
}
