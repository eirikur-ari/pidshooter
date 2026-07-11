//go:build integration

package service

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	require.NoError(t, svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0}))
	require.NotNil(t, store.Saved)
	require.Len(t, store.Saved.Scores, 1)
	entry := store.Saved.Scores[0]
	assert.Equal(t, 0, entry.Kills)
	assert.Greater(t, entry.Duration, 0.0)
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

	require.NoError(t, err)
	assert.Contains(t, stderr, "could not load scores")
	assert.Nil(t, store.Saved)
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

	require.NoError(t, err)
	assert.Contains(t, stderr, "score not saved")
	assert.Contains(t, stderr, "disk full")
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
	assert.NoError(t, svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0}))
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
	assert.NoError(t, svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0}))
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
	assert.NoError(t, svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0}))
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
	assert.NoError(t, svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0}))
}

func TestIntegration_GameService_TimeLimitExpires(t *testing.T) {
	svc := newGameService(
		&fake.Process{Processes: []process.Info{{Pid: 102, Name: "target", Rss: 1024}}},
		&fake.Store{},
		fake.NewInputSource(),
	)

	start := time.Now()
	require.NoError(t, svc.Play(inbound.GamePlayConfig{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 1}))
	assert.LessOrEqual(t, time.Since(start), 3*time.Second, "game took too long to exit on time limit")
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
