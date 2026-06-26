package game

import (
	"fmt"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestNewTarget_WithinBounds(t *testing.T) {
	maxX, maxY := 80, 24
	e := NewTarget(testutil.NewFakeProcess(1234, "test", 1024), maxX, maxY)

	if e.Pid() != 1234 {
		t.Errorf("expected PID=1234, got %d", e.Pid())
	}
	if e.Name() != "test" {
		t.Errorf("expected Name=test, got %s", e.Name())
	}
	if e.Rss() != 1024 {
		t.Errorf("expected RSS=1024, got %d", e.Rss())
	}
	if e.State != Alive {
		t.Errorf("expected State=StateAlive, got %d", e.State)
	}

	label := fmt.Sprintf("[%d %s]", e.Pid(), e.Name())
	labelLen := len(label)

	spawnMaxX := maxX - labelLen - 1
	spawnMaxY := maxY - 2
	if e.PosX < 1 || int(e.PosX) > spawnMaxX {
		t.Errorf("X=%f out of bounds [1, %d]", e.PosX, spawnMaxX)
	}
	if e.PosY < 1 || int(e.PosY) > spawnMaxY {
		t.Errorf("Y=%f out of bounds [1, %d]", e.PosY, spawnMaxY)
	}
}

func TestNewTarget_SmallTerminal(t *testing.T) {
	e := NewTarget(testutil.NewFakeProcess(1, "x", 0), 5, 5)
	if e == nil {
		t.Fatal("expected non-nil entity")
	}
}

func TestTarget_Label_Alive(t *testing.T) {
	e := &Target{Info: testutil.NewFakeProcess(42, "bash", 0), State: Alive}
	expected := "[42 bash]"
	if got := e.Label(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestTarget_Label_Dead(t *testing.T) {
	e := &Target{Info: testutil.NewFakeProcess(42, "bash", 0), State: Dead}
	if got := e.Label(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestTarget_Label_Killing(t *testing.T) {
	e := &Target{Info: testutil.NewFakeProcess(42, "bash", 0), State: Killing, KillAnimFrame: 0}
	label := e.Label()
	if label == "" {
		t.Error("expected non-empty kill animation label")
	}
}

func TestTarget_Update_KillingState(t *testing.T) {
	e := &Target{
		Info:          testutil.NewFakeProcess(1, "x", 0),
		State:         Killing,
		KillAnimFrame: KillAnimFrames - 1,
	}

	e.Update(80, 24, 1.0)

	if e.State != Dead {
		t.Errorf("expected StateDead after last kill frame, got %d", e.State)
	}
}

func TestTarget_Update_DeadNoOp(t *testing.T) {
	e := &Target{
		Info:   testutil.NewFakeProcess(1, "x", 0),
		Motion: Motion{PosX: 10, PosY: 10, VelX: 1.0, VelY: 1.0},
		State:  Dead,
	}

	e.Update(80, 24, 1.0)

	if e.PosX != 10 || e.PosY != 10 {
		t.Error("dead entity should not move")
	}
}

func TestTarget_Contains(t *testing.T) {
	e := &Target{
		Info:   testutil.NewFakeProcess(42, "bash", 0),
		Motion: Motion{PosX: 10, PosY: 5},
		State:  Alive,
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

func TestTarget_Contains_NotAlive(t *testing.T) {
	e := &Target{
		Info:   testutil.NewFakeProcess(42, "bash", 0),
		Motion: Motion{PosX: 10, PosY: 5},
		State:  Killing,
	}

	if e.Contains(10, 5) {
		t.Error("non-alive entity should not contain anything")
	}
}

func TestTarget_StartKillAnim(t *testing.T) {
	e := &Target{Info: testutil.NewFakeProcess(1, "x", 0), State: Alive, KillAnimFrame: 5}
	e.StartKillAnim()

	if e.State != Killing {
		t.Errorf("expected StateKilling, got %d", e.State)
	}
	if e.KillAnimFrame != 0 {
		t.Errorf("expected KillAnimFrame=0, got %d", e.KillAnimFrame)
	}
}
