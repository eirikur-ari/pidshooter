package game

import "testing"

func TestNewVelocity(t *testing.T) {
	v := NewVelocity(2.0)
	if v.Speed() != 2.0 {
		t.Errorf("expected 2.0, got %f", v.Speed())
	}
}

func TestVelocity_Increase(t *testing.T) {
	v := NewVelocity(2.0)
	v.Increase()
	if v.Speed() != 2.5 {
		t.Errorf("expected 2.5, got %f", v.Speed())
	}
}

func TestVelocity_Decrease(t *testing.T) {
	v := NewVelocity(2.0)
	v.Decrease()
	if v.Speed() != 1.5 {
		t.Errorf("expected 1.5, got %f", v.Speed())
	}
}

func TestVelocity_IncreaseCapsAtMax(t *testing.T) {
	v := NewVelocity(4.8)
	v.Increase()
	if v.Speed() != MaxSpeed {
		t.Errorf("expected %f, got %f", MaxSpeed, v.Speed())
	}
	v.Increase()
	if v.Speed() != MaxSpeed {
		t.Errorf("expected speed to remain %f, got %f", MaxSpeed, v.Speed())
	}
}

func TestVelocity_DecreaseFloorsAtMin(t *testing.T) {
	v := NewVelocity(0.3)
	v.Decrease()
	if v.Speed() != MinSpeed {
		t.Errorf("expected %f, got %f", MinSpeed, v.Speed())
	}
	v.Decrease()
	if v.Speed() != MinSpeed {
		t.Errorf("expected speed to remain %f, got %f", MinSpeed, v.Speed())
	}
}
