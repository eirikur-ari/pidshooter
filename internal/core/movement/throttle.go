package movement

const (
	MinSpeed = 0.1 // minimum allowed speed multiplier
	MaxSpeed = 5.0 // maximum allowed speed multiplier
)

// Throttler is the interface through which external packages adjust the throttle.
type Throttler interface {
	Increase()
	Decrease()
}

// Throttle tracks the current movement speed with clamped increase/decrease steps.
type Throttle struct {
	speed float64
}

// NewThrottle returns a Throttle initialised to the given speed.
func NewThrottle(speed float64) Throttle {
	return Throttle{speed: speed}
}

// Speed returns the current speed.
func (v *Throttle) Speed() float64 { return v.speed }

// Increase adds one step, capped at the maximum.
func (v *Throttle) Increase() {
	v.speed += 0.5
	if v.speed > MaxSpeed {
		v.speed = MaxSpeed
	}
}

// Decrease subtracts one step, floored at the minimum.
func (v *Throttle) Decrease() {
	v.speed -= 0.5
	if v.speed < MinSpeed {
		v.speed = MinSpeed
	}
}
