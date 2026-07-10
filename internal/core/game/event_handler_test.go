package game

import (
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// newRunningGame returns a Game in the Running state with the given fields set.
func newRunningGame(fields *Game) *Game {
	fields.Start(0, 0)
	return fields
}

// --- HandleKey ---

func TestHandleKey_Quit(t *testing.T) {
	g := newRunningGame(&Game{})
	g.HandleKey('q')
	if g.IsRunning() {
		t.Error("expected running=false after 'q'")
	}
}

func TestHandleKey_QuitUppercase(t *testing.T) {
	g := newRunningGame(&Game{})
	g.HandleKey('Q')
	if g.IsRunning() {
		t.Error("expected running=false after 'Q'")
	}
}

func TestHandleKey_SpeedUp(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(2.0)})
	g.HandleKey('+')
	if g.velocity.Speed() != 2.5 {
		t.Errorf("expected speed=2.5, got %f", g.velocity.Speed())
	}
}

func TestHandleKey_SpeedDown(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(2.0)})
	g.HandleKey('-')
	if g.velocity.Speed() != 1.5 {
		t.Errorf("expected speed=1.5, got %f", g.velocity.Speed())
	}
}

func TestHandleKey_SpeedUpAlias(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(2.0)})
	g.HandleKey('=')
	if g.velocity.Speed() != 2.5 {
		t.Errorf("expected speed=2.5 with '=' alias, got %f", g.velocity.Speed())
	}
}

func TestHandleKey_SpeedDownAlias(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(2.0)})
	g.HandleKey('_')
	if g.velocity.Speed() != 1.5 {
		t.Errorf("expected speed=1.5 with '_' alias, got %f", g.velocity.Speed())
	}
}

func TestHandleKey_SpeedCapsAtMax(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(4.8)})
	g.HandleKey('+')
	if g.velocity.Speed() != 5.0 {
		t.Errorf("expected speed capped at 5.0, got %f", g.velocity.Speed())
	}
	g.HandleKey('+')
	if g.velocity.Speed() != 5.0 {
		t.Errorf("expected speed still 5.0, got %f", g.velocity.Speed())
	}
}

func TestHandleKey_SpeedCapsAtMin(t *testing.T) {
	g := newRunningGame(&Game{velocity: NewVelocity(0.3)})
	g.HandleKey('-')
	if g.velocity.Speed() != 0.1 {
		t.Errorf("expected speed capped at 0.1, got %f", g.velocity.Speed())
	}
}

func TestHandleKey_ConfirmYes_ReturnsTarget(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 4096}, State: Alive}
	g := newRunningGame(&Game{confirm: Confirmation{target: tgt}})

	target := g.HandleKey('y')

	if target == nil || target != tgt {
		t.Error("expected confirmed target to be returned")
	}
	if g.confirm.Pending() {
		t.Error("expected confirm cleared after 'y'")
	}
}

func TestHandleKey_ConfirmNo(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 0}, State: Alive}
	g := newRunningGame(&Game{confirm: Confirmation{target: tgt}})

	target := g.HandleKey('n')

	if target != nil {
		t.Error("expected no target after 'n'")
	}
	if g.confirm.Pending() {
		t.Error("expected confirm cleared after 'n'")
	}
	if tgt.State != Alive {
		t.Errorf("expected target still alive, got %d", tgt.State)
	}
}

func TestHandleKey_QCancelsConfirm(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 0}, State: Alive}
	g := newRunningGame(&Game{confirm: Confirmation{target: tgt}})

	g.HandleKey('q')

	if g.confirm.Pending() {
		t.Error("expected confirm cleared after 'q' during confirmation")
	}
	if !g.IsRunning() {
		t.Error("expected game still running (q cancels confirm, doesn't quit)")
	}
}

// --- HandleClick ---

func TestHandleClick_ReturnsTarget(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{tgt}}

	target := g.HandleClick(10, 5)

	if target == nil || target != tgt {
		t.Error("expected clicked target to be returned")
	}
}

func TestHandleClick_SetsConfirmingInConfirmMode(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{tgt}, confirm: NewConfirmation(true)}

	target := g.HandleClick(10, 5)

	if target != nil {
		t.Error("expected no target in confirm mode (should set confirming instead)")
	}
	if g.confirm.target != tgt {
		t.Error("expected target set as pending confirmation")
	}
	if tgt.State != Alive {
		t.Errorf("expected target still Alive in confirm mode, got %d", tgt.State)
	}
}

func TestHandleClick_NoOpWhenAlreadyConfirming(t *testing.T) {
	existing := &Target{Info: process.Info{Pid: 1, Name: "other", Rss: 0}, State: Alive}
	tgt := &Target{Info: process.Info{Pid: 2, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{tgt}, confirm: Confirmation{target: existing}}

	result := g.HandleClick(10, 5)

	if result != nil {
		t.Error("expected no target when already confirming")
	}
	if g.confirm.target != existing {
		t.Error("expected pending confirmation to remain unchanged")
	}
}

func TestHandleClick_NoOpOnMiss(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{tgt}}

	target := g.HandleClick(0, 0)

	if target != nil {
		t.Error("expected no target on miss")
	}
	if tgt.State != Alive {
		t.Error("expected target to remain alive on miss")
	}
}
