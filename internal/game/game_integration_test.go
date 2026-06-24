//go:build integration

package game_test

import (
	"testing"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/game"
	"github.com/eirikur-ari/pidshooter/internal/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
	"github.com/gdamore/tcell/v2"
)

func newSimGame(t *testing.T, processes []process.Info, timeLimit int) (*game.Game, tcell.SimulationScreen) {
	t.Helper()
	sim := tcell.NewSimulationScreen("")
	sim.SetSize(80, 24)
	g := game.New(processes, false, 2.0, timeLimit)
	g.UseScreen(sim)
	return g, sim
}

func TestPlay_QuitOnQ(t *testing.T) {
	procs := []process.Info{testutil.NewFakeProcess(100, "target", 1024)}
	g, sim := newSimGame(t, procs, 0)

	go func() {
		time.Sleep(50 * time.Millisecond)
		sim.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	}()

	if err := g.Play(0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlay_QuitOnEscape(t *testing.T) {
	procs := []process.Info{testutil.NewFakeProcess(101, "target", 1024)}
	g, sim := newSimGame(t, procs, 0)

	go func() {
		time.Sleep(50 * time.Millisecond)
		sim.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	}()

	if err := g.Play(0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlay_TimeLimitExpires(t *testing.T) {
	procs := []process.Info{testutil.NewFakeProcess(102, "target", 1024)}
	g, _ := newSimGame(t, procs, 1)

	start := time.Now()
	if err := g.Play(0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("game took too long to exit on time limit: %v", elapsed)
	}
}

func TestPlay_SessionStateAfterQuit(t *testing.T) {
	procs := []process.Info{testutil.NewFakeProcess(103, "target", 1024)}
	g, sim := newSimGame(t, procs, 0)

	go func() {
		time.Sleep(50 * time.Millisecond)
		sim.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	}()

	if err := g.Play(0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.Kills() != 0 {
		t.Errorf("expected 0 kills after immediate quit, got %d", g.Kills())
	}
	if g.StartTime().IsZero() {
		t.Error("expected StartTime to be recorded after Play")
	}
}
