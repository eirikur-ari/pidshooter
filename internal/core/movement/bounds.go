package movement

// Bounds holds the terminal dimensions that constrain target movement.
type Bounds struct {
	width, height int
}

// NewBounds returns a Bounds with the given terminal dimensions.
func NewBounds(width, height int) Bounds {
	return Bounds{width: width, height: height}
}

// bounce adjusts position and velocity so the target stays within the frame,
// reversing the relevant velocity component on each wall hit.
func (b Bounds) bounce(pos, vel *Vector, tagWidth float64) {
	b.bounceLeft(pos, vel)
	b.bounceRight(pos, vel, tagWidth)
	b.bounceTop(pos, vel)
	b.bounceBottom(pos, vel)
}

func (b Bounds) bounceLeft(pos, vel *Vector) {
	if pos.X < 0 {
		pos.X = 0
		if vel.X < 0 {
			vel.X = -vel.X
		}
	}
}

func (b Bounds) bounceRight(pos, vel *Vector, tagWidth float64) {
	bound := float64(b.width) - tagWidth
	if bound < 0 {
		bound = 0
	}
	if pos.X > bound {
		pos.X = bound
		if vel.X > 0 {
			vel.X = -vel.X
		}
	}
}

func (b Bounds) bounceTop(pos, vel *Vector) {
	if pos.Y < 1 { // reserve top row for the HUD
		pos.Y = 1
		if vel.Y < 0 {
			vel.Y = -vel.Y
		}
	}
}

func (b Bounds) bounceBottom(pos, vel *Vector) {
	bound := float64(b.height - 2) // reserve bottom row for status bar
	if bound < 1 {
		bound = 1 // don't fight bounceTop's reserved row on very short terminals
	}
	if pos.Y > bound {
		pos.Y = bound
		if vel.Y > 0 {
			vel.Y = -vel.Y
		}
	}
}
