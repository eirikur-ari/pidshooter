package game

import (
	"testing"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/domain/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

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

func TestKillTarget_TransitionsToKilling(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 2048}, State: Alive}
	g := &Game{killer: &fake.Killer{}}

	g.killTarget(e)

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

func TestKillTarget_PassesNameToKiller(t *testing.T) {
	fk := &fake.Killer{}
	e := &Target{Info: process.Info{Pid: 42, Name: "myapp", Rss: 0}, State: Alive}
	g := &Game{killer: fk}

	g.killTarget(e)

	if len(fk.KilledNames) != 1 || fk.KilledNames[0] != "myapp" {
		t.Errorf("expected Kill called with name %q, got %v", "myapp", fk.KilledNames)
	}
}

func TestKillTarget_NoOpWhenNotAlive(t *testing.T) {
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 2048}, State: Dead}
	g := &Game{killer: &fake.Killer{}}

	g.killTarget(e)

	if g.kills != 0 {
		t.Errorf("expected kills=0 for dead entity, got %d", g.kills)
	}
}

func TestUpdate_StopsWhenTimeLimitExpired(t *testing.T) {
	renderer := &stubRenderer{w: 80, h: 24}
	g := &Game{renderer: renderer, timeLimit: 1}
	g.running.Store(true)
	g.startTime = time.Now().Add(-2 * time.Second)

	g.update()

	if g.running.Load() {
		t.Error("expected game stopped when time limit expired")
	}
}

func TestUpdate_StopsWhenAllTargetsDead(t *testing.T) {
	renderer := &stubRenderer{w: 80, h: 24}
	e := &Target{Info: process.Info{Pid: 1, Name: "target", Rss: 0}, State: Dead}
	g := &Game{renderer: renderer, targets: []*Target{e}}
	g.running.Store(true)
	g.startTime = time.Now()

	g.update()

	if g.running.Load() {
		t.Error("expected game stopped when all targets dead")
	}
}

func TestRender_AliveTargetIncluded(t *testing.T) {
	renderer := &stubRenderer{w: 80, h: 24}
	e := &Target{Info: process.Info{Pid: 1, Name: "myapp", Rss: 1024}, Motion: Motion{PosX: 10, PosY: 5}, State: Alive}
	g := &Game{renderer: renderer, targets: []*Target{e}}

	g.render()

	if len(renderer.frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(renderer.frames))
	}
	tv := renderer.frames[0].Targets
	if len(tv) != 1 {
		t.Fatalf("expected 1 target view, got %d", len(tv))
	}
	if tv[0].X != 10 || tv[0].Y != 5 {
		t.Errorf("expected position (10,5), got (%d,%d)", tv[0].X, tv[0].Y)
	}
	if tv[0].Killing {
		t.Error("expected Killing=false for alive target")
	}
}

func TestRender_DeadTargetExcluded(t *testing.T) {
	renderer := &stubRenderer{w: 80, h: 24}
	e := &Target{Info: process.Info{Pid: 1, Name: "myapp", Rss: 1024}, State: Dead}
	g := &Game{renderer: renderer, targets: []*Target{e}}

	g.render()

	if len(renderer.frames[0].Targets) != 0 {
		t.Errorf("expected dead target excluded from frame, got %d targets", len(renderer.frames[0].Targets))
	}
}

func TestRender_KillingTargetMarked(t *testing.T) {
	renderer := &stubRenderer{w: 80, h: 24}
	e := &Target{Info: process.Info{Pid: 1, Name: "myapp", Rss: 1024}, State: Killing}
	g := &Game{renderer: renderer, targets: []*Target{e}}

	g.render()

	tv := renderer.frames[0].Targets
	if len(tv) != 1 || !tv[0].Killing {
		t.Error("expected Killing=true for killing target")
	}
}

func TestRender_HUDReflectsSession(t *testing.T) {
	renderer := &stubRenderer{w: 80, h: 24}
	g := &Game{renderer: renderer, targets: []*Target{}}
	g.kills = 3
	g.freedMem = 2048
	g.highScore = 10

	g.render()

	hud := renderer.frames[0].HUD
	if hud.Kills != 3 {
		t.Errorf("expected Kills=3, got %d", hud.Kills)
	}
	if hud.FreedMem != 2048 {
		t.Errorf("expected FreedMem=2048, got %d", hud.FreedMem)
	}
	if hud.HighScore != 10 {
		t.Errorf("expected HighScore=10, got %d", hud.HighScore)
	}
}

func TestRender_ConfirmStateInStatusBar(t *testing.T) {
	renderer := &stubRenderer{w: 80, h: 24}
	target := &Target{Info: process.Info{Pid: 42, Name: "suspect", Rss: 0}, State: Alive}
	g := &Game{renderer: renderer, targets: []*Target{target}, confirming: target}

	g.render()

	cs := renderer.frames[0].StatusBar.Confirming
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

func TestRender_StatusBarAliveCount(t *testing.T) {
	renderer := &stubRenderer{w: 80, h: 24}
	alive := &Target{Info: process.Info{Pid: 1, Name: "a", Rss: 0}, State: Alive}
	dead := &Target{Info: process.Info{Pid: 2, Name: "b", Rss: 0}, State: Dead}
	killing := &Target{Info: process.Info{Pid: 3, Name: "c", Rss: 0}, State: Killing}
	g := &Game{renderer: renderer, targets: []*Target{alive, dead, killing}}

	g.render()

	status := renderer.frames[0].StatusBar
	if status.Alive != 1 {
		t.Errorf("expected Alive=1, got %d", status.Alive)
	}
}

