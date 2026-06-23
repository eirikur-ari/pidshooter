package game

import (
	"testing"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/process"
	"github.com/gdamore/tcell/v2"
)

type fakeProcess struct {
	pid  int
	name string
	rss  int64
}

func (f *fakeProcess) Pid() int     { return f.pid }
func (f *fakeProcess) Name() string { return f.name }
func (f *fakeProcess) Rss() int64   { return f.rss }

func TestNew(t *testing.T) {
	processes := []process.Info{
		&fakeProcess{pid: 1, name: "a", rss: 100},
		&fakeProcess{pid: 2, name: "b", rss: 200},
	}

	g := New(processes, true, 3.5, 60)

	if !g.confirmMode {
		t.Error("expected confirmMode=true")
	}
	if g.speed != 3.5 {
		t.Errorf("expected speed=3.5, got %f", g.speed)
	}
	if g.timeLimit != 60 {
		t.Errorf("expected timeLimit=60, got %d", g.timeLimit)
	}
	if !g.running {
		t.Error("expected running=true")
	}
	if g.kills != 0 {
		t.Errorf("expected kills=0, got %d", g.kills)
	}
	if g.freedMem != 0 {
		t.Errorf("expected freedMem=0, got %d", g.freedMem)
	}
}

func TestHandleKeyPress_Quit(t *testing.T) {
	g := &Game{running: true}
	g.HandleKeyPress(0, 'q')
	if g.running {
		t.Error("expected running=false after 'q'")
	}
}

func TestHandleKeyPress_QuitUppercase(t *testing.T) {
	g := &Game{running: true}
	g.HandleKeyPress(0, 'Q')
	if g.running {
		t.Error("expected running=false after 'Q'")
	}
}

func TestHandleKeyPress_Escape(t *testing.T) {
	g := &Game{running: true}
	g.HandleKeyPress(tcell.KeyEscape, 0)
	if g.running {
		t.Error("expected running=false after Escape")
	}
}

func TestHandleKeyPress_SpeedUp(t *testing.T) {
	g := &Game{running: true, speed: 2.0}
	g.HandleKeyPress(0, '+')
	if g.speed != 2.5 {
		t.Errorf("expected speed=2.5, got %f", g.speed)
	}
}

func TestHandleKeyPress_SpeedDown(t *testing.T) {
	g := &Game{running: true, speed: 2.0}
	g.HandleKeyPress(0, '-')
	if g.speed != 1.5 {
		t.Errorf("expected speed=1.5, got %f", g.speed)
	}
}

func TestHandleKeyPress_SpeedCapsAtMax(t *testing.T) {
	g := &Game{running: true, speed: 4.8}
	g.HandleKeyPress(0, '+')
	if g.speed != 5.0 {
		t.Errorf("expected speed capped at 5.0, got %f", g.speed)
	}
	g.HandleKeyPress(0, '+')
	if g.speed != 5.0 {
		t.Errorf("expected speed still 5.0, got %f", g.speed)
	}
}

func TestHandleKeyPress_SpeedCapsAtMin(t *testing.T) {
	g := &Game{running: true, speed: 0.3}
	g.HandleKeyPress(0, '-')
	if g.speed != 0.1 {
		t.Errorf("expected speed capped at 0.1, got %f", g.speed)
	}
}

func TestHandleKeyPress_ConfirmYes(t *testing.T) {
	e := &Entity{Info: &fakeProcess{pid: 99, name: "target", rss: 4096}, State: Alive}
	g := &Game{running: true, confirming: e}

	g.HandleKeyPress(0, 'y')

	if g.confirming != nil {
		t.Error("expected confirming=nil after 'y'")
	}
	if e.State != Killing {
		t.Errorf("expected entity StateKilling, got %d", e.State)
	}
	if g.kills != 1 {
		t.Errorf("expected kills=1, got %d", g.kills)
	}
	if g.freedMem != 4096 {
		t.Errorf("expected freedMem=4096, got %d", g.freedMem)
	}
}

func TestHandleKeyPress_ConfirmNo(t *testing.T) {
	e := &Entity{Info: &fakeProcess{pid: 99, name: "target"}, State: Alive}
	g := &Game{running: true, confirming: e}

	g.HandleKeyPress(0, 'n')

	if g.confirming != nil {
		t.Error("expected confirming=nil after 'n'")
	}
	if e.State != Alive {
		t.Errorf("expected entity still alive, got %d", e.State)
	}
}

func TestHandleKeyPress_QCancelsConfirm(t *testing.T) {
	e := &Entity{Info: &fakeProcess{pid: 99, name: "target"}, State: Alive}
	g := &Game{running: true, confirming: e}

	g.HandleKeyPress(0, 'q')

	if g.confirming != nil {
		t.Error("expected confirming=nil after 'q' during confirmation")
	}
	if !g.running {
		t.Error("expected game still running (q cancels confirm, doesn't quit)")
	}
}

func TestGame_timeRemaining_WithinLimit(t *testing.T) {
	g := &Game{timeLimit: 60}
	g.startTime = time.Now()
	remaining := g.timeRemaining()
	if remaining <= 0 || remaining > 60*time.Second {
		t.Errorf("expected remaining in (0, 60s], got %v", remaining)
	}
}

func TestGame_timeRemaining_Expired(t *testing.T) {
	g := &Game{timeLimit: 1}
	g.startTime = time.Now().Add(-2 * time.Second)
	if got := g.timeRemaining(); got != 0 {
		t.Errorf("expected 0 after expiry, got %v", got)
	}
}