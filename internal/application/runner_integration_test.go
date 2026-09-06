//go:build integration

package application

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func newRunner(proc *fake.Process, store *fake.Store, events *fake.InputSource, logger *fake.Logger) *Runner {
	return NewRunner(proc, store, &fake.Renderer{}, events, logger)
}

func TestIntegrationRunnerHappyPath(t *testing.T) {
	events := fake.NewInputSource()
	events.Ch <- outbound.QuitEvent{}

	store := &fake.Store{}
	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024}}},
		store,
		events,
		&fake.Logger{},
	)

	err := r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})
	require.NoError(t, err)
	require.NotNil(t, store.Saved)
	require.Len(t, store.Saved.Scores, 1)
	entry := store.Saved.Scores[0]
	assert.Equal(t, 0, entry.Kills)
	assert.Greater(t, entry.Duration, 0.0)
}

func TestIntegrationRunnerGameFailsWhenRendererInitFails(t *testing.T) {
	logger := &fake.Logger{}
	r := NewRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 204, Name: "target", Rss: 1024}}},
		&fake.Store{},
		&fake.Renderer{InitErr: errors.New("terminal not available")},
		fake.NewInputSource(),
		logger,
	)

	err := r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeGameFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	require.Len(t, logger.Errors, 1)
	assert.Contains(t, logger.Errors[0], "terminal not available")
}

func TestIntegrationRunnerLoadErrorPrintsWarningAndSkipsSave(t *testing.T) {
	events := fake.NewInputSource()
	events.Ch <- outbound.QuitEvent{}

	store := &fake.Store{LoadErr: errors.New("json: invalid character")}
	logger := &fake.Logger{}
	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 201, Name: "target", Rss: 1024}}},
		store,
		events,
		logger,
	)

	err := r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})

	require.NoError(t, err)
	require.Len(t, logger.Warnings, 2)
	assert.Contains(t, logger.Warnings[0], "score not loaded")
	assert.Contains(t, logger.Warnings[1], "score not saved")
	assert.Contains(t, logger.Warnings[1], "score not loaded", "should report the load failure as the reason the save was skipped")
	assert.Nil(t, store.Saved)
}

func TestIntegrationRunnerSaveErrorPrintsWarning(t *testing.T) {
	events := fake.NewInputSource()
	events.Ch <- outbound.QuitEvent{}

	logger := &fake.Logger{}
	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 202, Name: "target", Rss: 1024}}},
		&fake.Store{SaveErr: errors.New("disk full")},
		events,
		logger,
	)

	err := r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})

	require.NoError(t, err)
	require.Len(t, logger.Warnings, 1)
	assert.Contains(t, logger.Warnings[0], "score not saved")
	assert.Contains(t, logger.Warnings[0], "disk full")
}

func TestIntegrationRunnerNotFoundStillSaves(t *testing.T) {
	events := fake.NewInputSource()
	events.Ch <- outbound.QuitEvent{}

	store := &fake.Store{LoadErr: outbound.NotFoundError{}}
	logger := &fake.Logger{}
	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 203, Name: "target", Rss: 1024}}},
		store,
		events,
		logger,
	)

	err := r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})

	require.NoError(t, err)
	require.NotNil(t, store.Saved, "a fresh (never-persisted) board should still be saved")
	require.Len(t, logger.Warnings, 1)
	assert.Contains(t, logger.Warnings[0], "score not loaded")
}

func TestIntegrationRunnerQuitOnQuitEvent(t *testing.T) {
	events := fake.NewInputSource()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Ch <- outbound.QuitEvent{}
	}()

	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 100, Name: "target", Rss: 1024}}},
		&fake.Store{},
		events,
		&fake.Logger{},
	)
	assert.NoError(t, r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0}))
}

func TestIntegrationRunnerTimeLimitExpires(t *testing.T) {
	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 102, Name: "target", Rss: 1024}}},
		&fake.Store{},
		fake.NewInputSource(),
		&fake.Logger{},
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
			&fake.Process{Infos: []outbound.ProcessInfo{{PID: pid, Name: "target", Rss: 1024}}},
			&fake.Store{},
			events,
			&fake.Logger{},
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
