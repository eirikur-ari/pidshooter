package movement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewThrottle(t *testing.T) {
	v := NewThrottle(2.0)
	assert.Equal(t, 2.0, v.Speed())
}

func TestThrottle_Increase(t *testing.T) {
	v := NewThrottle(2.0)
	v.Increase()
	assert.Equal(t, 2.5, v.Speed())
}

func TestThrottle_Decrease(t *testing.T) {
	v := NewThrottle(2.0)
	v.Decrease()
	assert.Equal(t, 1.5, v.Speed())
}

func TestThrottle_IncreaseCapsAtMax(t *testing.T) {
	v := NewThrottle(4.8)
	v.Increase()
	assert.Equal(t, MaxSpeed, v.Speed())
	v.Increase()
	assert.Equal(t, MaxSpeed, v.Speed(), "expected speed to remain %f", MaxSpeed)
}

func TestThrottle_DecreaseFloorsAtMin(t *testing.T) {
	v := NewThrottle(0.3)
	v.Decrease()
	assert.Equal(t, MinSpeed, v.Speed())
	v.Decrease()
	assert.Equal(t, MinSpeed, v.Speed(), "expected speed to remain %f", MinSpeed)
}
