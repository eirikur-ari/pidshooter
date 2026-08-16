package movement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewThrottle(t *testing.T) {
	v := NewThrottle(2.0)
	assert.Equal(t, 2.0, v.Speed())
}

func TestThrottleIncrease(t *testing.T) {
	v := NewThrottle(2.0)
	v.Increase()
	assert.Equal(t, 2.5, v.Speed())
}

func TestThrottleDecrease(t *testing.T) {
	v := NewThrottle(2.0)
	v.Decrease()
	assert.Equal(t, 1.5, v.Speed())
}

func TestThrottleIncreaseCapsAtMax(t *testing.T) {
	v := NewThrottle(4.8)
	v.Increase()
	assert.Equal(t, MaxSpeed, v.Speed())
	v.Increase()
	assert.Equal(t, MaxSpeed, v.Speed(), "expected speed to remain %f", MaxSpeed)
}

func TestThrottleDecreaseFloorsAtMin(t *testing.T) {
	v := NewThrottle(0.8)
	v.Decrease()
	assert.Equal(t, MinSpeed, v.Speed())
	v.Decrease()
	assert.Equal(t, MinSpeed, v.Speed(), "expected speed to remain %f", MinSpeed)
}
