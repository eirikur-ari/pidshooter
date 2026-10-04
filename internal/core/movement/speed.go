package movement

// speed groups a current movement speed with the lowest value it has reached.
type speed struct {
	current float64
	lowest  float64
}

// newSpeed returns a speed initialized to the given value.
func newSpeed(initial float64) speed {
	return speed{current: initial, lowest: initial}
}

// Current returns the current speed value.
func (s *speed) Current() float64 { return s.current }

// Lowest returns the lowest value reached so far, including the starting value.
func (s *speed) Lowest() float64 { return s.lowest }

// Set updates the current value and tracks the lowest value reached.
func (s *speed) Set(value float64) {
	s.current = value
	if value < s.lowest {
		s.lowest = value
	}
}
