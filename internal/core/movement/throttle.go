package movement

import (
	"fmt"
	"math"
)

const (
	MinSpeed     = 0.5 // minimum allowed speed multiplier
	MaxSpeed     = 5.0 // maximum allowed speed multiplier
	throttleStep = 0.5 // increment/decrement per throttle adjustment
)

// Throttle tracks the current movement speed with clamped increase/decrease steps.
type Throttle struct {
	speed speed
}

// NewThrottle returns a Throttle initialized to the given speed.
func NewThrottle(initial float64) *Throttle {
	return &Throttle{speed: newSpeed(initial)}
}

// Speed returns the current speed.
func (t *Throttle) Speed() float64 { return t.speed.Current() }

// LowestSpeed returns the lowest speed reached so far, including the starting speed.
func (t *Throttle) LowestSpeed() float64 { return t.speed.Lowest() }

// Increase adds one step, capped at the maximum.
func (t *Throttle) Increase() {
	t.speed.Set(min(t.speed.Current()+throttleStep, MaxSpeed))
}

// Decrease subtracts one step, floored at the minimum.
func (t *Throttle) Decrease() {
	t.speed.Set(max(t.speed.Current()-throttleStep, MinSpeed))
}

// ValidateSpeed reports an error if value is not a finite number in
// [MinSpeed, MaxSpeed].
func ValidateSpeed(value float64) error {
	if math.IsNaN(value) || value < MinSpeed || value > MaxSpeed {
		return fmt.Errorf("speed must be between %g and %g, got: %g", MinSpeed, MaxSpeed, value)
	}
	return nil
}
