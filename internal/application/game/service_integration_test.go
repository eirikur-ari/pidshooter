//go:build integration

package game

import (
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/capture"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func newIntegrationRunner(finder *fake.Finder, killer *fake.Killer, store *fake.Store, events *stubEventSource) *GameService {
	return NewGameService(finder, killer, store, &stubRenderer{}, events)
}

func TestGameService_HappyPath(t *testing.T) {
	events := newStubEventSource()
	events.ch <- game.KeyEvent{Ch: 'q'}

	store := &fake.Store{}
	svc := newIntegrationRunner(
		&fake.Finder{Processes: []process.Info{{Pid: 1, Name: "target", Rss: 1024}}},
		&fake.Killer{},
		store,
		events,
	)

	if err := svc.Play(contract.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0}); err != nil {
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

func TestGameService_SaveError_PrintsWarning(t *testing.T) {
	events := newStubEventSource()
	events.ch <- game.KeyEvent{Ch: 'q'}

	svc := newIntegrationRunner(
		&fake.Finder{Processes: []process.Info{{Pid: 1, Name: "target", Rss: 1024}}},
		&fake.Killer{},
		&fake.Store{SaveErr: errors.New("disk full")},
		events,
	)

	var err error
	stderr := capture.Stderr(func() {
		err = svc.Play(contract.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})
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

func TestGameService_QuitOnQ(t *testing.T) {
	events := newStubEventSource()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.ch <- game.KeyEvent{Ch: 'q'}
	}()

	svc := newIntegrationRunner(
		&fake.Finder{Processes: []process.Info{{Pid: 100, Name: "target", Rss: 1024}}},
		&fake.Killer{},
		&fake.Store{},
		events,
	)
	if err := svc.Play(contract.Config{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGameService_QuitOnEscape(t *testing.T) {
	events := newStubEventSource()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.ch <- game.KeyEvent{Key: game.KeyEscape}
	}()

	svc := newIntegrationRunner(
		&fake.Finder{Processes: []process.Info{{Pid: 101, Name: "target", Rss: 1024}}},
		&fake.Killer{},
		&fake.Store{},
		events,
	)
	if err := svc.Play(contract.Config{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGameService_QuitOnCtrlC(t *testing.T) {
	events := newStubEventSource()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.ch <- game.KeyEvent{Key: game.KeyCtrlC}
	}()

	svc := newIntegrationRunner(
		&fake.Finder{Processes: []process.Info{{Pid: 104, Name: "target", Rss: 1024}}},
		&fake.Killer{},
		&fake.Store{},
		events,
	)
	if err := svc.Play(contract.Config{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGameService_QuitOnCtrlZ(t *testing.T) {
	events := newStubEventSource()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.ch <- game.KeyEvent{Key: game.KeyCtrlZ}
	}()

	svc := newIntegrationRunner(
		&fake.Finder{Processes: []process.Info{{Pid: 105, Name: "target", Rss: 1024}}},
		&fake.Killer{},
		&fake.Store{},
		events,
	)
	if err := svc.Play(contract.Config{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGameService_TimeLimitExpires(t *testing.T) {
	svc := newIntegrationRunner(
		&fake.Finder{Processes: []process.Info{{Pid: 102, Name: "target", Rss: 1024}}},
		&fake.Killer{},
		&fake.Store{},
		newStubEventSource(),
	)

	start := time.Now()
	if err := svc.Play(contract.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("game took too long to exit on time limit: %v", elapsed)
	}
}

// TestGameService_SignalGoroutineDoesNotAccumulate verifies the signal goroutine
// started inside runLoop exits when Play returns, preventing goroutine leaks.
func TestGameService_SignalGoroutineDoesNotAccumulate(t *testing.T) {
	runGame := func(pid int) {
		events := newStubEventSource()
		go func() {
			time.Sleep(50 * time.Millisecond)
			events.ch <- game.KeyEvent{Ch: 'q'}
		}()
		svc := newIntegrationRunner(
			&fake.Finder{Processes: []process.Info{{Pid: pid, Name: "target", Rss: 1024}}},
			&fake.Killer{},
			&fake.Store{},
			events,
		)
		if err := svc.Play(contract.Config{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
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
