package movement

// Speed groups a current movement speed with the lowest value it has reached.
type Speed struct {
	current float64
	lowest  float64
}

// NewSpeed returns a Speed initialised to the given value.
func NewSpeed(v float64) Speed {
	return Speed{current: v, lowest: v}
}

// Current returns the current speed value.
func (s Speed) Current() float64 { return s.current }

// Lowest returns the lowest value reached so far, including the starting value.
func (s Speed) Lowest() float64 { return s.lowest }

// Set updates the current value and tracks the lowest value reached.
func (s *Speed) Set(v float64) {
	s.current = v
	if v < s.lowest {
		s.lowest = v
	}
}