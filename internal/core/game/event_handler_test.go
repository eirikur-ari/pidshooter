package game

import (
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// newRunningGame returns a Game with running=true and the given fields set.
func newRunningGame(fields *Game) *Game {
	fields.running.Store(true)
	return fields
}

// --- HandleKey ---

func TestHandleKey_Quit(t *testing.T) {
	g := newRunningGame(&Game{})
	g.HandleKey('q')
	if g.running.Load() {
		t.Error("expected running=false after 'q'")
	}
}

func TestHandleKey_QuitUppercase(t *testing.T) {
	g := newRunningGame(&Game{})
	g.HandleKey('Q')
	if g.running.Load() {
		t.Error("expected running=false after 'Q'")
	}
}

func TestHandleKey_SpeedUp(t *testing.T) {
	g := newRunningGame(&Game{speed: 2.0})
	g.HandleKey('+')
	if g.speed != 2.5 {
		t.Errorf("expected speed=2.5, got %f", g.speed)
	}
}

func TestHandleKey_SpeedDown(t *testing.T) {
	g := newRunningGame(&Game{speed: 2.0})
	g.HandleKey('-')
	if g.speed != 1.5 {
		t.Errorf("expected speed=1.5, got %f", g.speed)
	}
}

func TestHandleKey_SpeedUpAlias(t *testing.T) {
	g := newRunningGame(&Game{speed: 2.0})
	g.HandleKey('=')
	if g.speed != 2.5 {
		t.Errorf("expected speed=2.5 with '=' alias, got %f", g.speed)
	}
}

func TestHandleKey_SpeedDownAlias(t *testing.T) {
	g := newRunningGame(&Game{speed: 2.0})
	g.HandleKey('_')
	if g.speed != 1.5 {
		t.Errorf("expected speed=1.5 with '_' alias, got %f", g.speed)
	}
}

func TestHandleKey_SpeedCapsAtMax(t *testing.T) {
	g := newRunningGame(&Game{speed: 4.8})
	g.HandleKey('+')
	if g.speed != 5.0 {
		t.Errorf("expected speed capped at 5.0, got %f", g.speed)
	}
	g.HandleKey('+')
	if g.speed != 5.0 {
		t.Errorf("expected speed still 5.0, got %f", g.speed)
	}
}

func TestHandleKey_SpeedCapsAtMin(t *testing.T) {
	g := newRunningGame(&Game{speed: 0.3})
	g.HandleKey('-')
	if g.speed != 0.1 {
		t.Errorf("expected speed capped at 0.1, got %f", g.speed)
	}
}

func TestHandleKey_ConfirmYes_ReturnsKillRequest(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 4096}, State: Alive}
	g := newRunningGame(&Game{confirming: e})

	req := g.HandleKey('y')

	if req == nil || req.Target != e {
		t.Error("expected KillRequest with the confirming target")
	}
	if g.confirming != nil {
		t.Error("expected confirming=nil after 'y'")
	}
}

func TestHandleKey_ConfirmNo(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 0}, State: Alive}
	g := newRunningGame(&Game{confirming: e})

	req := g.HandleKey('n')

	if req != nil {
		t.Error("expected no KillRequest after 'n'")
	}
	if g.confirming != nil {
		t.Error("expected confirming=nil after 'n'")
	}
	if e.State != Alive {
		t.Errorf("expected entity still alive, got %d", e.State)
	}
}

func TestHandleKey_QCancelsConfirm(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 0}, State: Alive}
	g := newRunningGame(&Game{confirming: e})

	g.HandleKey('q')

	if g.confirming != nil {
		t.Error("expected confirming=nil after 'q' during confirmation")
	}
	if !g.running.Load() {
		t.Error("expected game still running (q cancels confirm, doesn't quit)")
	}
}

// --- HandleClick ---

func TestHandleClick_ReturnsKillRequest(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{e}}

	req := g.HandleClick(10, 5)

	if req == nil || req.Target != e {
		t.Error("expected KillRequest for the clicked target")
	}
}

func TestHandleClick_SetsConfirmingInConfirmMode(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{e}, confirmMode: true}

	req := g.HandleClick(10, 5)

	if req != nil {
		t.Error("expected no KillRequest in confirm mode (should set confirming instead)")
	}
	if g.confirming != e {
		t.Error("expected entity set as confirming")
	}
	if e.State != Alive {
		t.Errorf("expected entity still Alive in confirm mode, got %d", e.State)
	}
}

func TestHandleClick_NoOpWhenAlreadyConfirming(t *testing.T) {
	existing := &Target{Info: process.Info{Pid: 1, Name: "other", Rss: 0}, State: Alive}
	target := &Target{Info: process.Info{Pid: 2, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{target}, confirming: existing}

	req := g.HandleClick(10, 5)

	if req != nil {
		t.Error("expected no KillRequest when already confirming")
	}
	if g.confirming != existing {
		t.Error("expected confirming to remain unchanged")
	}
}

func TestHandleClick_NoOpOnMiss(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{e}}

	req := g.HandleClick(0, 0)

	if req != nil {
		t.Error("expected no KillRequest on miss")
	}
	if e.State != Alive {
		t.Error("expected entity to remain alive on miss")
	}
}
