package game

// FrameVector is a 2D vector used for both position and velocity within the game frame.
type FrameVector struct {
	X, Y float64
}

// Apply moves pos by this vector scaled by speed.
func (v FrameVector) Apply(pos *FrameVector, speed float64) {
	pos.X += v.X * speed
	pos.Y += v.Y * speed
}
