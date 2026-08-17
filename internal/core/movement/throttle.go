package movement

const (
	MinSpeed  = 0.5 // minimum allowed speed multiplier (one full SpeedStep)
	MaxSpeed  = 5.0 // maximum allowed speed multiplier
	SpeedStep = 0.5 // increment/decrement per throttle adjustment
)

// Throttle tracks the current movement speed with clamped increase/decrease steps.
type Throttle struct {
	speed Speed
}

// NewThrottle returns a Throttle initialised to the given speed.
func NewThrottle(speed float64) *Throttle {
	return &Throttle{speed: NewSpeed(speed)}
}

// Speed returns the current speed.
func (t *Throttle) Speed() float64 { return t.speed.Current() }

// LowestSpeed returns the lowest speed reached so far, including the starting speed.
func (t *Throttle) LowestSpeed() float64 { return t.speed.Lowest() }

// Increase adds one step, capped at the maximum.
func (t *Throttle) Increase() {
	v := t.speed.Current() + SpeedStep
	if v > MaxSpeed {
		v = MaxSpeed
	}
	t.speed.Set(v)
}

// Decrease subtracts one step, floored at the minimum.
func (t *Throttle) Decrease() {
	v := t.speed.Current() - SpeedStep
	if v < MinSpeed {
		v = MinSpeed
	}
	t.speed.Set(v)
}
