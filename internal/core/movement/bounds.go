package movement

// WindowSize is a pair of dimensions bounding the area within which
// movement is constrained.
type WindowSize struct {
	Width, Height int
}

// ChromeSize is the number of rows reserved at the top and bottom edges of
// the window that movement must keep clear.
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

// bounce adjusts position and velocity so the position stays within the
// bounds, reversing the relevant velocity component on each wall hit. width
// is the horizontal extent of what is moving.
func (b Bounds) bounce(position, velocity *Vector, width int) {
	b.bounceLeft(position, velocity)
	b.bounceRight(position, velocity, width)
	b.bounceTop(position, velocity)
	b.bounceBottom(position, velocity)
}

func (b Bounds) bounceLeft(position, velocity *Vector) {
	if position.X < 0 {
		position.X = 0
		if velocity.X < 0 {
			velocity.X = -velocity.X
		}
	}
}

func (b Bounds) bounceRight(position, velocity *Vector, width int) {
	bound := float64(max(b.window.Width-width, 0))
	if position.X > bound {
		position.X = bound
		if velocity.X > 0 {
			velocity.X = -velocity.X
		}
	}
}

func (b Bounds) bounceTop(position, velocity *Vector) {
	bound := float64(b.chrome.Top)
	if position.Y < bound {
		position.Y = bound
		if velocity.Y < 0 {
			velocity.Y = -velocity.Y
		}
	}
}

func (b Bounds) bounceBottom(position, velocity *Vector) {
	const lastRowIndexOffset = 1 // b.window.Height is a row count; row indices are zero-based
	bound := float64(b.window.Height - lastRowIndexOffset - b.chrome.Bottom)
	if bound < float64(b.chrome.Top) {
		bound = float64(b.chrome.Top) // don't fight bounceTop's reserved rows on very short windows
	}
	if position.Y > bound {
		position.Y = bound
		if velocity.Y > 0 {
			velocity.Y = -velocity.Y
		}
	}
}
