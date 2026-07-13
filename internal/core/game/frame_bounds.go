package game

// FrameBounds holds the terminal dimensions that constrain target movement.
type FrameBounds struct {
	W, H int
}

// Bounce adjusts pos and vel so the target stays within the frame,
// reversing the relevant velocity component on each wall hit.
func (b FrameBounds) Bounce(pos, vel *FrameVector, width float64) {
	b.bounceLeft(pos, vel)
	b.bounceRight(pos, vel, width)
	b.bounceTop(pos, vel)
	b.bounceBottom(pos, vel)
}

func (b FrameBounds) bounceLeft(pos, vel *FrameVector) {
	if pos.X < 0 {
		pos.X = 0
		vel.X = -vel.X
	}
}

func (b FrameBounds) bounceRight(pos, vel *FrameVector, width float64) {
	bound := float64(b.W) - width
	if bound < 0 {
		bound = 0
	}
	if pos.X > bound {
		pos.X = bound
		vel.X = -vel.X
	}
}

func (b FrameBounds) bounceTop(pos, vel *FrameVector) {
	if pos.Y < 0 {
		pos.Y = 0
		vel.Y = -vel.Y
	}
}

func (b FrameBounds) bounceBottom(pos, vel *FrameVector) {
	bound := float64(b.H - 2) // reserve bottom row for status bar
	if bound < 0 {
		bound = 0
	}
	if pos.Y > bound {
		pos.Y = bound
		vel.Y = -vel.Y
	}
}
