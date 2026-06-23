package game

import "testing"

func TestNewMotion_WithinBounds(t *testing.T) {
	maxX, maxY, labelLen := 80, 24, 10
	for i := 0; i < 20; i++ {
		m := newMotion(maxX, maxY, labelLen)
		spawnMaxX := maxX - labelLen - 1
		spawnMaxY := maxY - 2
		if m.PosX < 1 || int(m.PosX) > spawnMaxX {
			t.Errorf("PosX=%f out of bounds [1, %d]", m.PosX, spawnMaxX)
		}
		if m.PosY < 1 || int(m.PosY) > spawnMaxY {
			t.Errorf("PosY=%f out of bounds [1, %d]", m.PosY, spawnMaxY)
		}
	}
}

func TestNewMotion_SmallTerminal(t *testing.T) {
	m := newMotion(5, 5, 10)
	if m.PosX < 1 {
		t.Errorf("PosX=%f should be >= 1", m.PosX)
	}
	if m.PosY < 1 {
		t.Errorf("PosY=%f should be >= 1", m.PosY)
	}
}

func TestMotion_Update_BounceLeft(t *testing.T) {
	m := Motion{PosX: 0, PosY: 5, VelX: -1.0, VelY: 0}
	m.Update(80, 24, 3, 1.0)
	if m.PosX < 0 {
		t.Errorf("PosX should not be negative, got %f", m.PosX)
	}
	if m.VelX < 0 {
		t.Errorf("VelX should be positive after left bounce, got %f", m.VelX)
	}
}

func TestMotion_Update_BounceRight(t *testing.T) {
	m := Motion{PosX: 78, PosY: 5, VelX: 2.0, VelY: 0}
	labelLen := 3.0
	m.Update(80, 24, labelLen, 1.0)
	rightBound := 80.0 - labelLen
	if m.PosX > rightBound {
		t.Errorf("PosX should not exceed right bound %f, got %f", rightBound, m.PosX)
	}
	if m.VelX > 0 {
		t.Errorf("VelX should be negative after right bounce, got %f", m.VelX)
	}
}

func TestMotion_Update_BounceTop(t *testing.T) {
	m := Motion{PosX: 5, PosY: 0, VelX: 0, VelY: -1.0}
	m.Update(80, 24, 3, 1.0)
	if m.PosY < 0 {
		t.Errorf("PosY should not be negative, got %f", m.PosY)
	}
	if m.VelY < 0 {
		t.Errorf("VelY should be positive after top bounce, got %f", m.VelY)
	}
}

func TestMotion_Update_BounceBottom(t *testing.T) {
	m := Motion{PosX: 5, PosY: 23, VelX: 0, VelY: 2.0}
	m.Update(80, 24, 3, 1.0)
	bottomBound := float64(24 - 2)
	if m.PosY > bottomBound {
		t.Errorf("PosY should not exceed bottom bound %f, got %f", bottomBound, m.PosY)
	}
	if m.VelY > 0 {
		t.Errorf("VelY should be negative after bottom bounce, got %f", m.VelY)
	}
}

func TestMotion_Update_SpeedMultiplier(t *testing.T) {
	m := Motion{PosX: 40, PosY: 12, VelX: 1.0, VelY: 0.5}
	startX, startY := m.PosX, m.PosY
	m.Update(80, 24, 3, 3.0)
	if m.PosX != startX+1.0*3.0 {
		t.Errorf("expected PosX=%f, got %f", startX+3.0, m.PosX)
	}
	if m.PosY != startY+0.5*3.0 {
		t.Errorf("expected PosY=%f, got %f", startY+1.5, m.PosY)
	}
}
