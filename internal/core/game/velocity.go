package game

const (
	MinSpeed = 0.1 // minimum allowed speed multiplier
	MaxSpeed = 5.0 // maximum allowed speed multiplier
)

// Speeder is the interface through which external packages adjust velocity.
type Speeder interface {
	Increase()
	Decrease()
}

// Velocity tracks the current movement speed with clamped increase/decrease steps.
type Velocity struct {
	speed float64
}

// NewVelocity returns a Velocity initialised to the given speed.
func NewVelocity(speed float64) Velocity {
	return Velocity{speed: speed}
}

// Speed returns the current speed.
func (v *Velocity) Speed() float64 { return v.speed }

// Increase adds one step, capped at the maximum.
func (v *Velocity) Increase() {
	v.speed += 0.5
	if v.speed > MaxSpeed {
		v.speed = MaxSpeed
	}
}

// Decrease subtracts one step, floored at the minimum.
func (v *Velocity) Decrease() {
	v.speed -= 0.5
	if v.speed < MinSpeed {
		v.speed = MinSpeed
	}
}
