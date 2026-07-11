package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVelocity(t *testing.T) {
	v := NewVelocity(2.0)
	assert.Equal(t, 2.0, v.Speed())
}

func TestVelocity_Increase(t *testing.T) {
	v := NewVelocity(2.0)
	v.Increase()
	assert.Equal(t, 2.5, v.Speed())
}

func TestVelocity_Decrease(t *testing.T) {
	v := NewVelocity(2.0)
	v.Decrease()
	assert.Equal(t, 1.5, v.Speed())
}

func TestVelocity_IncreaseCapsAtMax(t *testing.T) {
	v := NewVelocity(4.8)
	v.Increase()
	assert.Equal(t, MaxSpeed, v.Speed())
	v.Increase()
	assert.Equal(t, MaxSpeed, v.Speed(), "expected speed to remain %f", MaxSpeed)
}

func TestVelocity_DecreaseFloorsAtMin(t *testing.T) {
	v := NewVelocity(0.3)
	v.Decrease()
	assert.Equal(t, MinSpeed, v.Speed())
	v.Decrease()
	assert.Equal(t, MinSpeed, v.Speed(), "expected speed to remain %f", MinSpeed)
}
