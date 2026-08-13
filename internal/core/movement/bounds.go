package movement

// Bounds is the interface through which a moving entity is kept within the frame.
type Bounds interface {
	// Bounce adjusts position and velocity so the target stays within the frame,
	// reversing the relevant velocity component on each wall hit.
	Bounce(pos, vel *Vector, tagWidth float64)

	// Width returns the terminal width in columns.
	Width() int

	// Height returns the terminal height in rows.
	Height() int
}

// bounds holds the terminal dimensions that constrain target movement.
type bounds struct {
	width, height int
}

// NewBounds returns a Bounds with the given terminal dimensions.
func NewBounds(width, height int) Bounds {
	return bounds{width: width, height: height}
}

// Width returns the terminal width in columns.
func (b bounds) Width() int { return b.width }

// Height returns the terminal height in rows.
func (b bounds) Height() int { return b.height }

func (b bounds) Bounce(pos, vel *Vector, tagWidth float64) {
	b.bounceLeft(pos, vel)
	b.bounceRight(pos, vel, tagWidth)
	b.bounceTop(pos, vel)
	b.bounceBottom(pos, vel)
}

func (b bounds) bounceLeft(pos, vel *Vector) {
	if pos.X < 0 {
		pos.X = 0
		if vel.X < 0 {
			vel.X = -vel.X
		}
	}
}

func (b bounds) bounceRight(pos, vel *Vector, tagWidth float64) {
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

func (b bounds) bounceTop(pos, vel *Vector) {
	if pos.Y < 0 {
		pos.Y = 0
		if vel.Y < 0 {
			vel.Y = -vel.Y
		}
	}
}

func (b bounds) bounceBottom(pos, vel *Vector) {
	bound := float64(b.height - 2) // reserve bottom row for status bar
	if bound < 0 {
		bound = 0
	}
	if pos.Y > bound {
		pos.Y = bound
		if vel.Y > 0 {
			vel.Y = -vel.Y
		}
	}
}
