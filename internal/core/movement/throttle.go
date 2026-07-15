package movement

const (
	MinSpeed = 0.1 // minimum allowed speed multiplier
	MaxSpeed = 5.0 // maximum allowed speed multiplier
)

// Throttle tracks the current movement speed with clamped increase/decrease steps.
type Throttle struct {
	speed float64
}

// NewThrottle returns a Throttle initialised to the given speed.
func NewThrottle(speed float64) Throttle {
	return Throttle{speed: speed}
}

// Speed returns the current speed.
func (t *Throttle) Speed() float64 { return t.speed }

// Increase adds one step, capped at the maximum.
func (t *Throttle) Increase() {
	t.speed += 0.5
	if t.speed > MaxSpeed {
		t.speed = MaxSpeed
	}
}

// Decrease subtracts one step, floored at the minimum.
func (t *Throttle) Decrease() {
	t.speed -= 0.5
	if t.speed < MinSpeed {
		t.speed = MinSpeed
	}
}
