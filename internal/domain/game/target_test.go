package game

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func TestNewTarget_WithinBounds(t *testing.T) {
	maxX, maxY := 80, 24
	e := NewTarget(fake.NewProcess(1234, "test", 1024), maxX, maxY)

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
	e := NewTarget(fake.NewProcess(1, "x", 0), 5, 5)
	if e == nil {
		t.Fatal("expected non-nil entity")
	}
}

func TestTarget_Label_Alive(t *testing.T) {
	e := &Target{Info: fake.NewProcess(42, "bash", 0), State: Alive}
	expected := "[42 bash]"
	if got := e.Label(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestTarget_Label_Dead(t *testing.T) {
	e := &Target{Info: fake.NewProcess(42, "bash", 0), State: Dead}
	if got := e.Label(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestTarget_Label_Killing(t *testing.T) {
	e := &Target{Info: fake.NewProcess(42, "bash", 0), State: Killing, KillAnimFrame: 0}
	label := e.Label()
	if label == "" {
		t.Error("expected non-empty kill animation label")
	}
}

func TestTarget_Update_MultiByteRightWall(t *testing.T) {
	// "[42 café]" is 9 runes but 10 UTF-8 bytes.
	// With the byte-count bug, rightBound = maxX - 10 = 70.
	// With the fix, rightBound = maxX - 9 = 71.
	// Place the entity at PosX=70.5 moving right at speed=1. After one update:
	//   fix:  new PosX = 71.0 — at the correct boundary, no bounce yet.
	//   bug:  new PosX > 70 → bounce, VelX flips negative.
	e := &Target{
		Info:   fake.NewProcess(42, "café", 0),
		Motion: Motion{PosX: 70.5, PosY: 5, VelX: 0.5, VelY: 0},
		State:  Alive,
	}
	e.Update(80, 24, 1.0)
	if e.VelX < 0 {
		t.Errorf("entity bounced prematurely at right wall — byte-count bug in Update?"+
			" PosX=%.1f VelX=%.1f", e.PosX, e.VelX)
	}
}

func TestTarget_Update_KillingState(t *testing.T) {
	e := &Target{
		Info:          fake.NewProcess(1, "x", 0),
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
		Info:   fake.NewProcess(1, "x", 0),
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
		Info:   fake.NewProcess(42, "bash", 0),
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

func TestTarget_Contains_MultiByteProcessName(t *testing.T) {
	// "café" is 5 UTF-8 bytes but 4 runes → label "[42 café]" is 10 bytes, 9 runes.
	// With the byte-count bug, Contains over-counts by 1 and accepts column 19 as a hit.
	e := &Target{
		Info:   fake.NewProcess(42, "café", 0),
		Motion: Motion{PosX: 10, PosY: 5},
		State:  Alive,
	}
	label := e.Label()
	runeCount := utf8.RuneCountInString(label)

	pastEnd := 10 + runeCount
	if e.Contains(pastEnd, 5) {
		t.Errorf("Contains(%d, 5) should be false for label %q (rune count %d) — byte-count bug?",
			pastEnd, label, runeCount)
	}
}

func TestTarget_Contains_NotAlive(t *testing.T) {
	e := &Target{
		Info:   fake.NewProcess(42, "bash", 0),
		Motion: Motion{PosX: 10, PosY: 5},
		State:  Killing,
	}

	if e.Contains(10, 5) {
		t.Error("non-alive entity should not contain anything")
	}
}

func TestTarget_StartKillAnim(t *testing.T) {
	e := &Target{Info: fake.NewProcess(1, "x", 0), State: Alive, KillAnimFrame: 5}
	e.StartKillAnim()

	if e.State != Killing {
		t.Errorf("expected StateKilling, got %d", e.State)
	}
	if e.KillAnimFrame != 0 {
		t.Errorf("expected KillAnimFrame=0, got %d", e.KillAnimFrame)
	}
}
