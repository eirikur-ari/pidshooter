package game

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/eirikur-ari/pidshooter/internal/domain/process"
)

func TestNewTarget_WithinBounds(t *testing.T) {
	maxX, maxY := 80, 24
	e := NewTarget(process.Info{Pid: 1234, Name: "test", Rss: 1024}, maxX, maxY)

	if e.Pid != 1234 {
		t.Errorf("expected PID=1234, got %d", e.Pid)
	}
	if e.Name != "test" {
		t.Errorf("expected Name=test, got %s", e.Name)
	}
	if e.Rss != 1024 {
		t.Errorf("expected RSS=1024, got %d", e.Rss)
	}
	if e.State != Alive {
		t.Errorf("expected State=Alive, got %d", e.State)
	}

	label := fmt.Sprintf("[%d %s]", e.Pid, e.Name)
	labelLen := len(label)
	spawnMaxX := maxX - labelLen - 1
	spawnMaxY := maxY - 2
	if e.Position.X < 1 || int(e.Position.X) > spawnMaxX {
		t.Errorf("Position.X=%f out of bounds [1, %d]", e.Position.X, spawnMaxX)
	}
	if e.Position.Y < 1 || int(e.Position.Y) > spawnMaxY {
		t.Errorf("Position.Y=%f out of bounds [1, %d]", e.Position.Y, spawnMaxY)
	}
}

func TestNewTarget_SmallTerminal(t *testing.T) {
	e := NewTarget(process.Info{Pid: 1, Name: "xxx", Rss: 0}, 5, 5)
	if e == nil {
		t.Fatal("expected non-nil entity")
	}
}

func TestTarget_Label_Alive(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 42, Name: "bash", Rss: 0}, State: Alive}
	expected := "[42 bash]"
	if got := e.Label(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestTarget_Label_Dead(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 42, Name: "bash", Rss: 0}, State: Dead}
	if got := e.Label(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestTarget_Label_Killing(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 42, Name: "bash", Rss: 0}, State: Killing, KillAnimFrame: 0}
	if label := e.Label(); label == "" {
		t.Error("expected non-empty kill animation label")
	}
}

func TestTarget_Update_KillingState(t *testing.T) {
	e := &Target{
		Info:          process.Info{Pid: 1, Name: "xxx", Rss: 0},
		State:         Killing,
		KillAnimFrame: KillAnimFrames - 1,
	}
	e.Update(80, 24, 1.0)
	if e.State != Dead {
		t.Errorf("expected Dead after last kill frame, got %d", e.State)
	}
}

func TestTarget_Update_DeadNoOp(t *testing.T) {
	e := &Target{
		Info:     process.Info{Pid: 1, Name: "xxx", Rss: 0},
		Position: Vector{X: 10, Y: 10},
		Velocity: Vector{X: 1.0, Y: 1.0},
		State:    Dead,
	}
	e.Update(80, 24, 1.0)
	if e.Position.X != 10 || e.Position.Y != 10 {
		t.Error("dead entity should not move")
	}
}

func TestTarget_Update_BounceLeft(t *testing.T) {
	e := &Target{
		Info:     process.Info{Pid: 1, Name: "x"},
		Position: Vector{X: 0, Y: 5},
		Velocity: Vector{X: -1.0, Y: 0},
		State:    Alive,
	}
	e.Update(80, 24, 1.0)
	if e.Position.X < 0 {
		t.Errorf("Position.X should not be negative after left bounce, got %f", e.Position.X)
	}
	if e.Velocity.X < 0 {
		t.Errorf("Velocity.X should be positive after left bounce, got %f", e.Velocity.X)
	}
}

func TestTarget_Update_BounceRight(t *testing.T) {
	// label "[1 x]" = 5 chars → rightBound = 80-5 = 75
	e := &Target{
		Info:     process.Info{Pid: 1, Name: "x"},
		Position: Vector{X: 75, Y: 5},
		Velocity: Vector{X: 2.0, Y: 0},
		State:    Alive,
	}
	e.Update(80, 24, 1.0)
	rightBound := 75.0
	if e.Position.X > rightBound {
		t.Errorf("Position.X should not exceed right bound %f after right bounce, got %f", rightBound, e.Position.X)
	}
	if e.Velocity.X > 0 {
		t.Errorf("Velocity.X should be negative after right bounce, got %f", e.Velocity.X)
	}
}

func TestTarget_Update_BounceTop(t *testing.T) {
	e := &Target{
		Info:     process.Info{Pid: 1, Name: "x"},
		Position: Vector{X: 5, Y: 0},
		Velocity: Vector{X: 0, Y: -1.0},
		State:    Alive,
	}
	e.Update(80, 24, 1.0)
	if e.Position.Y < 0 {
		t.Errorf("Position.Y should not be negative after top bounce, got %f", e.Position.Y)
	}
	if e.Velocity.Y < 0 {
		t.Errorf("Velocity.Y should be positive after top bounce, got %f", e.Velocity.Y)
	}
}

func TestTarget_Update_BounceBottom(t *testing.T) {
	e := &Target{
		Info:     process.Info{Pid: 1, Name: "x"},
		Position: Vector{X: 5, Y: 23},
		Velocity: Vector{X: 0, Y: 2.0},
		State:    Alive,
	}
	e.Update(80, 24, 1.0)
	bottomBound := float64(24 - 2)
	if e.Position.Y > bottomBound {
		t.Errorf("Position.Y should not exceed bottom bound %f after bottom bounce, got %f", bottomBound, e.Position.Y)
	}
	if e.Velocity.Y > 0 {
		t.Errorf("Velocity.Y should be negative after bottom bounce, got %f", e.Velocity.Y)
	}
}

func TestTarget_Update_SpeedMultiplier(t *testing.T) {
	// label "[1 x]" = 5 chars; at (40,10) with speed=3 there is no wall bounce.
	e := &Target{
		Info:     process.Info{Pid: 1, Name: "x"},
		Position: Vector{X: 40, Y: 10},
		Velocity: Vector{X: 1.0, Y: 0.5},
		State:    Alive,
	}
	e.Update(80, 24, 3.0)
	if e.Position.X != 43.0 {
		t.Errorf("expected Position.X=43.0, got %f", e.Position.X)
	}
	if e.Position.Y != 11.5 {
		t.Errorf("expected Position.Y=11.5, got %f", e.Position.Y)
	}
}

func TestTarget_Update_MultiByteRightWall(t *testing.T) {
	// "[42 café]" is 9 runes but 10 UTF-8 bytes.
	// With the byte-count bug, rightBound = maxX - 10 = 70.
	// With the fix, rightBound = maxX - 9 = 71.
	// Place the entity at Position.X=70.5 moving right at speed=1. After one update:
	//   fix:  new Position.X = 71.0 — at the correct boundary, no bounce yet.
	//   bug:  new Position.X > 70 → bounce, Velocity.X flips negative.
	e := &Target{
		Info:     process.Info{Pid: 42, Name: "café", Rss: 0},
		Position: Vector{X: 70.5, Y: 5},
		Velocity: Vector{X: 0.5, Y: 0},
		State:    Alive,
	}
	e.Update(80, 24, 1.0)
	if e.Velocity.X < 0 {
		t.Errorf("entity bounced prematurely at right wall — byte-count bug in Update?"+
			" Position.X=%.1f Velocity.X=%.1f", e.Position.X, e.Velocity.X)
	}
}

func TestTarget_Contains(t *testing.T) {
	e := &Target{
		Info:     process.Info{Pid: 42, Name: "bash", Rss: 0},
		Position: Vector{X: 10, Y: 5},
		State:    Alive,
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
		Info:     process.Info{Pid: 42, Name: "café", Rss: 0},
		Position: Vector{X: 10, Y: 5},
		State:    Alive,
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
		Info:     process.Info{Pid: 42, Name: "bash", Rss: 0},
		Position: Vector{X: 10, Y: 5},
		State:    Killing,
	}
	if e.Contains(10, 5) {
		t.Error("non-alive entity should not contain anything")
	}
}

func TestTarget_StartKillAnim(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "xxx", Rss: 0}, State: Alive, KillAnimFrame: 5}
	e.StartKillAnim()
	if e.State != Killing {
		t.Errorf("expected Killing, got %d", e.State)
	}
	if e.KillAnimFrame != 0 {
		t.Errorf("expected KillAnimFrame=0, got %d", e.KillAnimFrame)
	}
}
