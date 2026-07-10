package game

// Velocity tracks the current movement speed with clamped increase/decrease steps.
type Velocity struct {
	value float64
}

// NewVelocity returns a Velocity initialised to v.
func NewVelocity(v float64) Velocity {
	return Velocity{value: v}
}

// Speed returns the current speed.
func (v Velocity) Speed() float64 { return v.value }

// Increase adds one step, capped at the maximum.
func (v *Velocity) Increase() {
	v.value += 0.5
	if v.value > 5.0 {
		v.value = 5.0
	}
}

// Decrease subtracts one step, floored at the minimum.
func (v *Velocity) Decrease() {
	v.value -= 0.5
	if v.value < 0.1 {
		v.value = 0.1
	}
}
