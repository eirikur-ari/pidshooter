package movement

// WindowSize is a pair of dimensions bounding the area within which
// movement is constrained.
type WindowSize struct {
	Width, Height int
}

// ChromeSize is the number of rows reserved at the top and bottom edges of
// that area that movement must keep clear.
type ChromeSize struct {
	Top, Bottom int
}

// Bounds holds the dimensions that constrain movement, and the rows
// reserved at the top and bottom edges that movement must keep clear.
type Bounds struct {
	window WindowSize
	chrome ChromeSize
}

// NewBounds returns a Bounds with the given window dimensions and reserved
// chrome rows.
func NewBounds(window WindowSize, chrome ChromeSize) Bounds {
	return Bounds{window: window, chrome: chrome}
}

// Update returns a copy of the bounds with its window dimensions replaced,
// keeping the same reserved chrome rows.
func (b Bounds) Update(window WindowSize) Bounds {
	return Bounds{window: window, chrome: b.chrome}
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
	bound := float64(b.window.Width) - tagWidth
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
	bound := float64(b.chrome.Top)
	if pos.Y < bound {
		pos.Y = bound
		if vel.Y < 0 {
			vel.Y = -vel.Y
		}
	}
}

func (b Bounds) bounceBottom(pos, vel *Vector) {
	const lastRowIndexOffset = 1 // b.window.Height is a row count; row indices are zero-based
	bound := float64(b.window.Height - lastRowIndexOffset - b.chrome.Bottom)
	if bound < float64(b.chrome.Top) {
		bound = float64(b.chrome.Top) // don't fight bounceTop's reserved rows on very short windows
	}
	if pos.Y > bound {
		pos.Y = bound
		if vel.Y > 0 {
			vel.Y = -vel.Y
		}
	}
}
