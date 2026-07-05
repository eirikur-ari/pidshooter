package game

import (
	"fmt"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"
	"github.com/eirikur-ari/pidshooter/internal/domain/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// newRunningGame returns a Game with running=true and the given fields set.
// atomic.Bool cannot be set in a struct literal, so helpers use this constructor.
func newRunningGame(fields *Game) *Game {
	fields.running.Store(true)
	return fields
}

// --- handleKeyPress ---

func TestHandleKeyPress_Quit(t *testing.T) {
	g := newRunningGame(&Game{})
	g.handleKeyPress(0, 'q')
	if g.running.Load() {
		t.Error("expected running=false after 'q'")
	}
}

func TestHandleKeyPress_QuitUppercase(t *testing.T) {
	g := newRunningGame(&Game{})
	g.handleKeyPress(0, 'Q')
	if g.running.Load() {
		t.Error("expected running=false after 'Q'")
	}
}

func TestHandleKeyPress_Escape(t *testing.T) {
	g := newRunningGame(&Game{})
	g.handleKeyPress(driven.KeyEscape, 0)
	if g.running.Load() {
		t.Error("expected running=false after Escape")
	}
}

func TestHandleKeyPress_CtrlC(t *testing.T) {
	g := newRunningGame(&Game{})
	g.handleKeyPress(driven.KeyCtrlC, 0)
	if g.running.Load() {
		t.Error("expected running=false after CtrlC")
	}
}

func TestHandleKeyPress_CtrlZ(t *testing.T) {
	g := newRunningGame(&Game{})
	g.handleKeyPress(driven.KeyCtrlZ, 0)
	if g.running.Load() {
		t.Error("expected running=false after CtrlZ")
	}
}

func TestHandleKeyPress_SpeedUp(t *testing.T) {
	g := newRunningGame(&Game{speed: 2.0})
	g.handleKeyPress(0, '+')
	if g.speed != 2.5 {
		t.Errorf("expected speed=2.5, got %f", g.speed)
	}
}

func TestHandleKeyPress_SpeedDown(t *testing.T) {
	g := newRunningGame(&Game{speed: 2.0})
	g.handleKeyPress(0, '-')
	if g.speed != 1.5 {
		t.Errorf("expected speed=1.5, got %f", g.speed)
	}
}

func TestHandleKeyPress_SpeedCapsAtMax(t *testing.T) {
	g := newRunningGame(&Game{speed: 4.8})
	g.handleKeyPress(0, '+')
	if g.speed != 5.0 {
		t.Errorf("expected speed capped at 5.0, got %f", g.speed)
	}
	g.handleKeyPress(0, '+')
	if g.speed != 5.0 {
		t.Errorf("expected speed still 5.0, got %f", g.speed)
	}
}

func TestHandleKeyPress_SpeedCapsAtMin(t *testing.T) {
	g := newRunningGame(&Game{speed: 0.3})
	g.handleKeyPress(0, '-')
	if g.speed != 0.1 {
		t.Errorf("expected speed capped at 0.1, got %f", g.speed)
	}
}

func TestHandleKeyPress_ConfirmYes(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 4096}, State: Alive}
	fk := &fake.Killer{}
	g := newRunningGame(&Game{confirming: e, killer: fk})

	g.handleKeyPress(0, 'y')

	if g.confirming != nil {
		t.Error("expected confirming=nil after 'y'")
	}
	if e.State != Killing {
		t.Errorf("expected entity Killing, got %d", e.State)
	}
	if g.kills != 1 {
		t.Errorf("expected kills=1, got %d", g.kills)
	}
	if g.freedMem != 4096 {
		t.Errorf("expected freedMem=4096, got %d", g.freedMem)
	}
	if len(fk.KilledPIDs) != 1 || fk.KilledPIDs[0] != 1 {
		t.Errorf("expected Kill(1) called once, got %v", fk.KilledPIDs)
	}
}

func TestHandleKeyPress_ConfirmYes_KillError(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 4096}, State: Alive}
	fk := &fake.Killer{Err: fmt.Errorf("process already finished")}
	g := newRunningGame(&Game{confirming: e, killer: fk})

	g.handleKeyPress(0, 'y')

	if e.State != Alive {
		t.Errorf("expected entity still Alive after failed kill, got %d", e.State)
	}
	if g.kills != 0 {
		t.Errorf("expected kills=0 after failed kill, got %d", g.kills)
	}
	if g.freedMem != 0 {
		t.Errorf("expected freedMem=0 after failed kill, got %d", g.freedMem)
	}
}

func TestHandleKeyPress_ConfirmNo(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 0}, State: Alive}
	g := newRunningGame(&Game{confirming: e})

	g.handleKeyPress(0, 'n')

	if g.confirming != nil {
		t.Error("expected confirming=nil after 'n'")
	}
	if e.State != Alive {
		t.Errorf("expected entity still alive, got %d", e.State)
	}
}

func TestHandleKeyPress_QCancelsConfirm(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 0}, State: Alive}
	g := newRunningGame(&Game{confirming: e})

	g.handleKeyPress(0, 'q')

	if g.confirming != nil {
		t.Error("expected confirming=nil after 'q' during confirmation")
	}
	if !g.running.Load() {
		t.Error("expected game still running (q cancels confirm, doesn't quit)")
	}
}

// --- handleMouseClick ---

func TestHandleMouseClick_KillsTargetOnClick(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	fk := &fake.Killer{}
	g := &Game{targets: []*Target{e}, killer: fk}

	g.handleMouseClick(10, 5)

	if e.State != Killing {
		t.Errorf("expected entity Killing after click, got %d", e.State)
	}
	if g.kills != 1 {
		t.Errorf("expected kills=1, got %d", g.kills)
	}
	if len(fk.KilledPIDs) != 1 || fk.KilledPIDs[0] != 1 {
		t.Errorf("expected Kill(1) called once, got %v", fk.KilledPIDs)
	}
}

func TestHandleMouseClick_KillError(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{e}, killer: &fake.Killer{Err: fmt.Errorf("no such process")}}

	g.handleMouseClick(10, 5)

	if e.State != Alive {
		t.Errorf("expected entity still Alive after failed kill, got %d", e.State)
	}
	if g.kills != 0 {
		t.Errorf("expected kills=0 after failed kill, got %d", g.kills)
	}
}

func TestHandleMouseClick_SetsConfirmingInConfirmMode(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{e}, confirmMode: true}

	g.handleMouseClick(10, 5)

	if g.confirming != e {
		t.Error("expected entity set as confirming")
	}
	if e.State != Alive {
		t.Errorf("expected entity still Alive in confirm mode, got %d", e.State)
	}
}

func TestHandleMouseClick_NoOpWhenAlreadyConfirming(t *testing.T) {
	existing := &Target{Info: process.Info{Pid: 1, Name: "other", Rss: 0}, State: Alive}
	target := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{target}, confirming: existing}

	g.handleMouseClick(10, 5)

	if g.confirming != existing {
		t.Error("expected confirming to remain unchanged")
	}
	if target.State != Alive {
		t.Error("expected target to remain alive")
	}
}

func TestHandleMouseClick_NoOpOnMiss(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{e}}

	g.handleMouseClick(0, 0)

	if e.State != Alive {
		t.Error("expected entity to remain alive on miss")
	}
	if g.kills != 0 {
		t.Errorf("expected kills=0, got %d", g.kills)
	}
}
