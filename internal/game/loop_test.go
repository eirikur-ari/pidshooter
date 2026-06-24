package game

import (
	"testing"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/testutil"
	"github.com/gdamore/tcell/v2"
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
	e := &Target{Info: testutil.NewFakeProcess(1, "target", 2048), State: Alive}
	g := &Game{}

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

func TestKillTarget_NoOpWhenNotAlive(t *testing.T) {
	e := &Target{Info: testutil.NewFakeProcess(1, "target", 2048), State: Dead}
	g := &Game{}

	g.killTarget(e)

	if g.kills != 0 {
		t.Errorf("expected kills=0 for dead entity, got %d", g.kills)
	}
}

func TestUpdate_StopsWhenAllTargetsDead(t *testing.T) {
	sim := tcell.NewSimulationScreen("")
	sim.SetSize(80, 24)
	_ = sim.Init()

	e := &Target{Info: testutil.NewFakeProcess(1, "target", 0), State: Dead}
	g := &Game{screen: sim, running: true, targets: []*Target{e}}
	g.startTime = time.Now()

	g.update()

	if g.running {
		t.Error("expected game stopped when all targets dead")
	}
}