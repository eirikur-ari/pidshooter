//go:build integration

package service

import (
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/event"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/capture"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func newGameService(proc *fake.Process, store *fake.Store, events *fake.InputSource) *GameService {
	return NewGameService(proc, store, &fake.Renderer{}, events)
}

func TestIntegration_GameService_HappyPath(t *testing.T) {
	events := fake.NewInputSource()
	events.Ch <- event.KeyEvent{Ch: 'q'}

	store := &fake.Store{}
	svc := newGameService(
		&fake.Process{Processes: []process.Info{{Pid: 1, Name: "target", Rss: 1024}}},
		store,
		events,
	)

	if err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if store.Saved == nil {
		t.Fatal("expected Save to be called, got nil")
	}
	if len(store.Saved.Scores) != 1 {
		t.Fatalf("expected 1 score entry, got %d", len(store.Saved.Scores))
	}
	entry := store.Saved.Scores[0]
	if entry.Kills != 0 {
		t.Errorf("expected 0 kills after immediate quit, got %d", entry.Kills)
	}
	if entry.Duration <= 0 {
		t.Errorf("expected positive duration, got %f", entry.Duration)
	}
}

func TestIntegration_GameService_LoadError_PrintsWarningAndSkipsSave(t *testing.T) {
	events := fake.NewInputSource()
	events.Ch <- event.KeyEvent{Ch: 'q'}

	store := &fake.Store{LoadErr: errors.New("json: invalid character")}
	svc := newGameService(
		&fake.Process{Processes: []process.Info{{Pid: 1, Name: "target", Rss: 1024}}},
		store,
		events,
	)

	var err error
	stderr := capture.Stderr(func() {
		err = svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})
	})

	if err != nil {
		t.Fatalf("expected Play to return nil, got %v", err)
	}
	if !strings.Contains(stderr, "could not load scores") {
		t.Errorf("expected stderr warning about load failure, got: %q", stderr)
	}
	if store.Saved != nil {
		t.Error("expected Save to be skipped after load failure, but it was called")
	}
}

func TestIntegration_GameService_SaveError_PrintsWarning(t *testing.T) {
	events := fake.NewInputSource()
	events.Ch <- event.KeyEvent{Ch: 'q'}

	svc := newGameService(
		&fake.Process{Processes: []process.Info{{Pid: 1, Name: "target", Rss: 1024}}},
		&fake.Store{SaveErr: errors.New("disk full")},
		events,
	)

	var err error
	stderr := capture.Stderr(func() {
		err = svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})
	})

	if err != nil {
		t.Fatalf("expected Play to return nil, got %v", err)
	}
	if !strings.Contains(stderr, "score not saved") {
		t.Errorf("expected stderr warning about save failure, got: %q", stderr)
	}
	if !strings.Contains(stderr, "disk full") {
		t.Errorf("expected stderr to include underlying error, got: %q", stderr)
	}
}

func TestIntegration_GameService_QuitOnQ(t *testing.T) {
	events := fake.NewInputSource()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Ch <- event.KeyEvent{Ch: 'q'}
	}()

	svc := newGameService(
		&fake.Process{Processes: []process.Info{{Pid: 100, Name: "target", Rss: 1024}}},
		&fake.Store{},
		events,
	)
	if err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntegration_GameService_QuitOnEscape(t *testing.T) {
	events := fake.NewInputSource()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Ch <- event.KeyEvent{Key: event.KeyEscape}
	}()

	svc := newGameService(
		&fake.Process{Processes: []process.Info{{Pid: 101, Name: "target", Rss: 1024}}},
		&fake.Store{},
		events,
	)
	if err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntegration_GameService_QuitOnCtrlC(t *testing.T) {
	events := fake.NewInputSource()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Ch <- event.KeyEvent{Key: event.KeyCtrlC}
	}()

	svc := newGameService(
		&fake.Process{Processes: []process.Info{{Pid: 104, Name: "target", Rss: 1024}}},
		&fake.Store{},
		events,
	)
	if err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntegration_GameService_QuitOnCtrlZ(t *testing.T) {
	events := fake.NewInputSource()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Ch <- event.KeyEvent{Key: event.KeyCtrlZ}
	}()

	svc := newGameService(
		&fake.Process{Processes: []process.Info{{Pid: 105, Name: "target", Rss: 1024}}},
		&fake.Store{},
		events,
	)
	if err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntegration_GameService_TimeLimitExpires(t *testing.T) {
	svc := newGameService(
		&fake.Process{Processes: []process.Info{{Pid: 102, Name: "target", Rss: 1024}}},
		&fake.Store{},
		fake.NewInputSource(),
	)

	start := time.Now()
	if err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("game took too long to exit on time limit: %v", elapsed)
	}
}

// TestIntegration_GameService_SignalGoroutineDoesNotAccumulate verifies the signal goroutine
// started inside runLoop exits when Play returns, preventing goroutine leaks.
func TestIntegration_GameService_SignalGoroutineDoesNotAccumulate(t *testing.T) {
	runGame := func(pid int) {
		events := fake.NewInputSource()
		go func() {
			time.Sleep(50 * time.Millisecond)
			events.Ch <- event.KeyEvent{Ch: 'q'}
		}()
		svc := newGameService(
			&fake.Process{Processes: []process.Info{{Pid: pid, Name: "target", Rss: 1024}}},
			&fake.Store{},
			events,
		)
		if err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
			t.Fatalf("pid %d: unexpected error: %v", pid, err)
		}
	}

	// Warm up os/signal's lazily-created background goroutine.
	runGame(299)
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	before := runtime.NumGoroutine()

	for i := 0; i < 3; i++ {
		runGame(300 + i)
	}

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
