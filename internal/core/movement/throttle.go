package movement

const (
	MinSpeed  = 0.5 // minimum allowed speed multiplier (one full SpeedStep)
	MaxSpeed  = 5.0 // maximum allowed speed multiplier
	SpeedStep = 0.5 // increment/decrement per throttle adjustment
)

// Throttler is the interface through which external packages adjust the movement throttle.
type Throttler interface {
	Speed() float64
	Increase()
	Decrease()
}

// throttle tracks the current movement speed with clamped increase/decrease steps.
type throttle struct {
	speed float64
}

// NewThrottle returns a Throttler initialised to the given speed.
func NewThrottle(speed float64) Throttler {
	return &throttle{speed: speed}
}

// Speed returns the current speed.
func (t *throttle) Speed() float64 { return t.speed }

// Increase adds one step, capped at the maximum.
func (t *throttle) Increase() {
	t.speed += SpeedStep
	if t.speed > MaxSpeed {
		t.speed = MaxSpeed
	}
}

// Decrease subtracts one step, floored at the minimum.
func (t *throttle) Decrease() {
	t.speed -= SpeedStep
	if t.speed < MinSpeed {
		t.speed = MinSpeed
	}
}
