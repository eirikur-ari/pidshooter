//go:build integration

package application

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil/capture"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func newRunner(proc *fake.Process, store *fake.Store, events *fake.InputSource) *Runner {
	return NewRunner(proc, store, &fake.Renderer{}, events)
}

func TestIntegrationRunnerHappyPath(t *testing.T) {
	events := fake.NewInputSource()
	events.Ch <- outbound.QuitEvent{}

	store := &fake.Store{}
	r := newRunner(
		&fake.Process{Processes: []outbound.ProcessInfo{{Pid: 200, Name: "target", Rss: 1024}}},
		store,
		events,
	)

	require.NoError(t, r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0}))
	require.NotNil(t, store.Saved)
	require.Len(t, store.Saved.Scores, 1)
	entry := store.Saved.Scores[0]
	assert.Equal(t, 0, entry.Kills)
	assert.Greater(t, entry.Duration, 0.0)
}

func TestIntegrationRunnerLoadErrorPrintsWarningAndSkipsSave(t *testing.T) {
	events := fake.NewInputSource()
	events.Ch <- outbound.QuitEvent{}

	store := &fake.Store{LoadErr: errors.New("json: invalid character")}
	r := newRunner(
		&fake.Process{Processes: []outbound.ProcessInfo{{Pid: 201, Name: "target", Rss: 1024}}},
		store,
		events,
	)

	var err error
	stderr := capture.Stderr(func() {
		err = r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})
	})

	require.NoError(t, err)
	assert.Contains(t, stderr, "could not load scores")
	assert.Nil(t, store.Saved)
}

func TestIntegrationRunnerSaveErrorPrintsWarning(t *testing.T) {
	events := fake.NewInputSource()
	events.Ch <- outbound.QuitEvent{}

	r := newRunner(
		&fake.Process{Processes: []outbound.ProcessInfo{{Pid: 202, Name: "target", Rss: 1024}}},
		&fake.Store{SaveErr: errors.New("disk full")},
		events,
	)

	var err error
	stderr := capture.Stderr(func() {
		err = r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})
	})

	require.NoError(t, err)
	assert.Contains(t, stderr, "score not saved")
	assert.Contains(t, stderr, "disk full")
}

func TestIntegrationRunnerQuitOnQuitEvent(t *testing.T) {
	events := fake.NewInputSource()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Ch <- outbound.QuitEvent{}
	}()

	r := newRunner(
		&fake.Process{Processes: []outbound.ProcessInfo{{Pid: 100, Name: "target", Rss: 1024}}},
		&fake.Store{},
		events,
	)
	assert.NoError(t, r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0}))
}

func TestIntegrationRunnerTimeLimitExpires(t *testing.T) {
	r := newRunner(
		&fake.Process{Processes: []outbound.ProcessInfo{{Pid: 102, Name: "target", Rss: 1024}}},
		&fake.Store{},
		fake.NewInputSource(),
	)

	start := time.Now()
	require.NoError(t, r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 1}))
	assert.LessOrEqual(t, time.Since(start), 3*time.Second, "game took too long to exit on time limit")
}

// TestIntegrationRunnerSignalGoroutineDoesNotAccumulate verifies the signal goroutine
// started inside runLoop exits when Run returns, preventing goroutine leaks.
func TestIntegrationRunnerSignalGoroutineDoesNotAccumulate(t *testing.T) {
	runGame := func(pid int) {
		events := fake.NewInputSource()
		go func() {
			time.Sleep(50 * time.Millisecond)
			events.Ch <- outbound.QuitEvent{}
		}()
		r := newRunner(
			&fake.Process{Processes: []outbound.ProcessInfo{{Pid: pid, Name: "target", Rss: 1024}}},
			&fake.Store{},
			events,
		)
		if err := r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
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
