package game

import (
	"fmt"
	"testing"
)

func TestNewEntity_WithinBounds(t *testing.T) {
	maxX, maxY := 80, 24
	e := NewEntity(1234, "test", 1024, maxX, maxY)

	if e.PID != 1234 {
		t.Errorf("expected PID=1234, got %d", e.PID)
	}
	if e.Name != "test" {
		t.Errorf("expected Name=test, got %s", e.Name)
	}
	if e.RSS != 1024 {
		t.Errorf("expected RSS=1024, got %d", e.RSS)
	}
	if e.State != StateAlive {
		t.Errorf("expected State=StateAlive, got %d", e.State)
	}

	label := fmt.Sprintf("[%d %s]", e.PID, e.Name)
	labelLen := len(label)

	spawnMaxX := maxX - labelLen - 1
	spawnMaxY := maxY - 2
	if e.X < 1 || int(e.X) > spawnMaxX {
		t.Errorf("X=%f out of bounds [1, %d]", e.X, spawnMaxX)
	}
	if e.Y < 1 || int(e.Y) > spawnMaxY {
		t.Errorf("Y=%f out of bounds [1, %d]", e.Y, spawnMaxY)
	}
}

func TestNewEntity_SmallTerminal(t *testing.T) {
	e := NewEntity(1, "x", 0, 5, 5)
	if e == nil {
		t.Fatal("expected non-nil entity")
	}
}

func TestEntity_Label_Alive(t *testing.T) {
	e := &Entity{PID: 42, Name: "bash", State: StateAlive}
	expected := "[42 bash]"
	if got := e.Label(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestEntity_Label_Dead(t *testing.T) {
	e := &Entity{PID: 42, Name: "bash", State: StateDead}
	if got := e.Label(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestEntity_Label_Killing(t *testing.T) {
	e := &Entity{PID: 42, Name: "bash", State: StateKilling, KillAnimFrame: 0}
	label := e.Label()
	if label == "" {
		t.Error("expected non-empty kill animation label")
	}
}

func TestEntity_Update_Bounce(t *testing.T) {
	e := &Entity{
		PID:   1,
		Name:  "x",
		State: StateAlive,
		X:     0, Y: 0,
		VelX: -1.0, VelY: -1.0,
	}

	e.Update(80, 24, 1.0)

	if e.X < 0 {
		t.Errorf("X should not be negative, got %f", e.X)
	}
	if e.Y < 0 {
		t.Errorf("Y should not be negative, got %f", e.Y)
	}
	if e.VelX < 0 {
		t.Errorf("VelX should be positive after left bounce, got %f", e.VelX)
	}
	if e.VelY < 0 {
		t.Errorf("VelY should be positive after top bounce, got %f", e.VelY)
	}
}

func TestEntity_Update_BounceRight(t *testing.T) {
	e := &Entity{
		PID:   1,
		Name:  "x",
		State: StateAlive,
		X:     78, Y: 10,
		VelX: 2.0, VelY: 0,
	}

	e.Update(80, 24, 1.0)

	labelLen := float64(len(e.Label()))
	rightBound := 80.0 - labelLen
	if e.X > rightBound {
		t.Errorf("X should not exceed right bound %f, got %f", rightBound, e.X)
	}
	if e.VelX > 0 {
		t.Errorf("VelX should be negative after right bounce, got %f", e.VelX)
	}
}

func TestEntity_Update_BounceBottom(t *testing.T) {
	e := &Entity{
		PID:   1,
		Name:  "x",
		State: StateAlive,
		X:     10, Y: 23,
		VelX: 0, VelY: 2.0,
	}

	e.Update(80, 24, 1.0)

	bottomBound := float64(24 - 2)
	if e.Y > bottomBound {
		t.Errorf("Y should not exceed bottom bound %f, got %f", bottomBound, e.Y)
	}
	if e.VelY > 0 {
		t.Errorf("VelY should be negative after bottom bounce, got %f", e.VelY)
	}
}

func TestEntity_Update_SpeedMultiplier(t *testing.T) {
	e := &Entity{
		PID:   1,
		Name:  "x",
		State: StateAlive,
		X:     40, Y: 12,
		VelX: 1.0, VelY: 0.5,
	}

	startX := e.X
	startY := e.Y
	e.Update(80, 24, 3.0)

	expectedX := startX + 1.0*3.0
	expectedY := startY + 0.5*3.0
	if e.X != expectedX {
		t.Errorf("expected X=%f, got %f", expectedX, e.X)
	}
	if e.Y != expectedY {
		t.Errorf("expected Y=%f, got %f", expectedY, e.Y)
	}
}

func TestEntity_Update_KillingState(t *testing.T) {
	e := &Entity{
		PID:           1,
		Name:          "x",
		State:         StateKilling,
		KillAnimFrame: KillAnimFrames - 1,
	}

	e.Update(80, 24, 1.0)

	if e.State != StateDead {
		t.Errorf("expected StateDead after last kill frame, got %d", e.State)
	}
}

func TestEntity_Update_DeadNoOp(t *testing.T) {
	e := &Entity{
		PID:   1,
		Name:  "x",
		State: StateDead,
		X:     10, Y: 10,
		VelX: 1.0, VelY: 1.0,
	}

	e.Update(80, 24, 1.0)

	if e.X != 10 || e.Y != 10 {
		t.Error("dead entity should not move")
	}
}

func TestEntity_Contains(t *testing.T) {
	e := &Entity{
		PID:   42,
		Name:  "bash",
		State: StateAlive,
		X:     10,
		Y:     5,
	}

	label := e.Label()
	labelLen := len(label)

	if !e.Contains(10, 5) {
		t.Error("expected Contains(10,5)=true")
	}
	if !e.Contains(10+labelLen-1, 5) {
		t.Error("expected Contains at last char=true")
	}
	if e.Contains(9, 5) {
		t.Error("expected Contains(9,5)=false")
	}
	if e.Contains(10+labelLen, 5) {
		t.Error("expected Contains past end=false")
	}
	if e.Contains(10, 4) {
		t.Error("expected Contains wrong row=false")
	}
}

func TestEntity_Contains_NotAlive(t *testing.T) {
	e := &Entity{
		PID:   42,
		Name:  "bash",
		State: StateKilling,
		X:     10,
		Y:     5,
	}

	if e.Contains(10, 5) {
		t.Error("non-alive entity should not contain anything")
	}
}

func TestEntity_StartKillAnim(t *testing.T) {
	e := &Entity{PID: 1, Name: "x", State: StateAlive, KillAnimFrame: 5}
	e.StartKillAnim()

	if e.State != StateKilling {
		t.Errorf("expected StateKilling, got %d", e.State)
	}
	if e.KillAnimFrame != 0 {
		t.Errorf("expected KillAnimFrame=0, got %d", e.KillAnimFrame)
	}
}
