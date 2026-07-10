package game

import (
	"testing"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestGame_timeRemaining_WithinLimit(t *testing.T) {
	g := &Game{cfg: Config{TimeLimit: 60}}
	g.startTime = time.Now()
	remaining := g.timeRemaining()
	if remaining <= 0 || remaining > 60*time.Second {
		t.Errorf("expected remaining in (0, 60s], got %v", remaining)
	}
}

func TestGame_timeRemaining_Expired(t *testing.T) {
	g := &Game{cfg: Config{TimeLimit: 1}}
	g.startTime = time.Now().Add(-2 * time.Second)
	if got := g.timeRemaining(); got != 0 {
		t.Errorf("expected 0 after expiry, got %v", got)
	}
}

func TestCompleteKill_TransitionsToKilling(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 2048}, State: Alive}
	g := &Game{}

	g.CompleteKill(e)

	if e.State != Killing {
		t.Errorf("expected entity Killing, got %d", e.State)
	}
	if g.kills != 1 {
		t.Errorf("expected kills=1, got %d", g.kills)
	}
	if g.freedMem != 2048 {
		t.Errorf("expected freedMem=2048, got %d", g.freedMem)
	}
}

func TestCompleteKill_NoOpWhenNotAlive(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 2048}, State: Dead}
	g := &Game{}

	g.CompleteKill(e)

	if g.kills != 0 {
		t.Errorf("expected kills=0 for dead entity, got %d", g.kills)
	}
}

func TestUpdate_StopsWhenTimeLimitExpired(t *testing.T) {
	g := &Game{cfg: Config{TimeLimit: 1}}
	g.running.Store(true)
	g.startTime = time.Now().Add(-2 * time.Second)

	g.Update(80, 24)

	if g.running.Load() {
		t.Error("expected game stopped when time limit expired")
	}
}

func TestUpdate_StopsWhenAllTargetsDead(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 0}, State: Dead}
	g := &Game{targets: []*Target{e}}
	g.running.Store(true)
	g.startTime = time.Now()

	g.Update(80, 24)

	if g.running.Load() {
		t.Error("expected game stopped when all targets dead")
	}
}

func TestFrame_AliveTargetIncluded(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "myapp", Rss: 1024}, Position: Vector{X: 10, Y: 5}, State: Alive}
	g := &Game{targets: []*Target{e}}

	frame := g.Frame()

	if len(frame.Targets) != 1 {
		t.Fatalf("expected 1 target view, got %d", len(frame.Targets))
	}
	tv := frame.Targets[0]
	if tv.X != 10 || tv.Y != 5 {
		t.Errorf("expected position (10,5), got (%d,%d)", tv.X, tv.Y)
	}
	if tv.Killing {
		t.Error("expected Killing=false for alive target")
	}
}

func TestFrame_DeadTargetExcluded(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "myapp", Rss: 1024}, State: Dead}
	g := &Game{targets: []*Target{e}}

	frame := g.Frame()

	if len(frame.Targets) != 0 {
		t.Errorf("expected dead target excluded from frame, got %d targets", len(frame.Targets))
	}
}

func TestFrame_KillingTargetMarked(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "myapp", Rss: 1024}, State: Killing}
	g := &Game{targets: []*Target{e}}

	frame := g.Frame()

	if len(frame.Targets) != 1 || !frame.Targets[0].Killing {
		t.Error("expected Killing=true for killing target")
	}
}

func TestFrame_HUDReflectsSession(t *testing.T) {
	g := &Game{targets: []*Target{}}
	g.kills = 3
	g.freedMem = 2048
	g.highScore = 10

	frame := g.Frame()

	if frame.HUD.Kills != 3 {
		t.Errorf("expected Kills=3, got %d", frame.HUD.Kills)
	}
	if frame.HUD.FreedMem != 2048 {
		t.Errorf("expected FreedMem=2048, got %d", frame.HUD.FreedMem)
	}
	if frame.HUD.HighScore != 10 {
		t.Errorf("expected HighScore=10, got %d", frame.HUD.HighScore)
	}
}

func TestFrame_ConfirmStateInStatusBar(t *testing.T) {
	target := &Target{Info: process.Info{Pid: 42, Name: "suspect", Rss: 0}, State: Alive}
	g := &Game{targets: []*Target{target}, confirming: target}

	frame := g.Frame()

	cs := frame.StatusBar.Confirming
	if cs == nil {
		t.Fatal("expected ConfirmState in status bar")
	}
	if cs.PID != 42 {
		t.Errorf("expected PID=42, got %d", cs.PID)
	}
	if cs.Name != "suspect" {
		t.Errorf("expected Name=suspect, got %q", cs.Name)
	}
}

func TestFrame_StatusBarAliveCount(t *testing.T) {
	alive := &Target{Info: process.Info{Pid: 1, Name: "a", Rss: 0}, State: Alive}
	dead := &Target{Info: process.Info{Pid: 2, Name: "b", Rss: 0}, State: Dead}
	killing := &Target{Info: process.Info{Pid: 3, Name: "c", Rss: 0}, State: Killing}
	g := &Game{targets: []*Target{alive, dead, killing}}

	frame := g.Frame()

	if frame.StatusBar.Alive != 1 {
		t.Errorf("expected Alive=1, got %d", frame.StatusBar.Alive)
	}
}
