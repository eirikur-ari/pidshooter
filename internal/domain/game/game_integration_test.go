//go:build integration

package game_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/domain/game"
	gamedriven "github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"
	procdriven "github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

type testRenderer struct{ w, h int }

func (r *testRenderer) Init() error           { return nil }
func (r *testRenderer) Cleanup()              {}
func (r *testRenderer) Size() (int, int)      { return r.w, r.h }
func (r *testRenderer) Render(_ gamedriven.Frame)  {}

type testEventSource struct{ ch chan gamedriven.InputEvent }

func newTestEventSource() *testEventSource {
	return &testEventSource{ch: make(chan gamedriven.InputEvent, 100)}
}
func (e *testEventSource) Events() <-chan gamedriven.InputEvent { return e.ch }
func (e *testEventSource) Send(ev gamedriven.InputEvent)        { e.ch <- ev }

func newTestGame(t *testing.T, processes []procdriven.Info, timeLimit int) (*game.Game, *testEventSource) {
	t.Helper()
	events := newTestEventSource()
	g := game.New(processes, false, 2.0, timeLimit,
		&fake.Killer{},
		&testRenderer{w: 80, h: 24},
		events,
	)
	return g, events
}

func TestPlay_QuitOnQ(t *testing.T) {
	procs := []procdriven.Info{fake.NewProcess(100, "target", 1024)}
	g, events := newTestGame(t, procs, 0)

	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Send(gamedriven.KeyEvent{Ch: 'q'})
	}()

	if err := g.Play(0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlay_QuitOnEscape(t *testing.T) {
	procs := []procdriven.Info{fake.NewProcess(101, "target", 1024)}
	g, events := newTestGame(t, procs, 0)

	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Send(gamedriven.KeyEvent{Key: gamedriven.KeyEscape})
	}()

	if err := g.Play(0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlay_TimeLimitExpires(t *testing.T) {
	procs := []procdriven.Info{fake.NewProcess(102, "target", 1024)}
	g, _ := newTestGame(t, procs, 1)

	start := time.Now()
	if err := g.Play(0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("game took too long to exit on time limit: %v", elapsed)
	}
}

// TestPlay_SignalGoroutineDoesNotAccumulate is a regression test for issue #7.
// It verifies that the signal goroutine started inside Play() exits when Play
// returns. Before the fix, signal.Stop did not close sigCh, so the goroutine
// stayed blocked on <-sigCh forever — one leaked goroutine per game session.
func TestPlay_SignalGoroutineDoesNotAccumulate(t *testing.T) {
	runGame := func(pid int) {
		procs := []procdriven.Info{fake.NewProcess(pid, "target", 1024)}
		g, events := newTestGame(t, procs, 0)
		go func() {
			time.Sleep(50 * time.Millisecond)
			events.Send(gamedriven.KeyEvent{Ch: 'q'})
		}()
		if err := g.Play(0); err != nil {
			t.Fatalf("pid %d: unexpected error: %v", pid, err)
		}
	}

	// Warm up os/signal's lazily-created background goroutine so it is
	// counted in the baseline and does not inflate later measurements.
	runGame(299)
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	before := runtime.NumGoroutine()

	for i := 0; i < 3; i++ {
		runGame(300 + i)
	}

	// Poll until goroutines return to baseline; exit is async after deferred close(done).
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("signal goroutines accumulated: want ≤%d goroutines after 3 games, got %d",
		before, runtime.NumGoroutine())
}

func TestPlay_SessionStateAfterQuit(t *testing.T) {
	procs := []procdriven.Info{fake.NewProcess(103, "target", 1024)}
	g, events := newTestGame(t, procs, 0)

	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Send(gamedriven.KeyEvent{Ch: 'q'})
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
