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
	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func newRunner(proc *fake.Process, store *fake.Store, events *fake.InputEventSource, logger *fake.Logger) *Runner {
	return NewRunner(proc, store, &fake.ScoreReporter{}, &fake.Renderer{}, events, logger)
}

func TestIntegrationRunnerRunIsSuccessful(t *testing.T) {
	events := fake.NewInputEventSource()
	events.Ch <- input.QuitEvent{}

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

func TestIntegrationRunnerRunReturnsErrorWhenRendererInitFails(t *testing.T) {
	logger := &fake.Logger{}
	r := NewRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 204, Name: "target", Rss: 1024}}},
		&fake.Store{},
		&fake.ScoreReporter{},
		&fake.Renderer{InitErr: errors.New("terminal not available")},
		fake.NewInputEventSource(),
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

func TestIntegrationRunnerRunPrintsWarningAndSkipsSaveWhenLoadingScoreBoardFails(t *testing.T) {
	events := fake.NewInputEventSource()
	events.Ch <- input.QuitEvent{}

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
	assert.Contains(t, logger.Warnings[0], "score board not loaded")
	assert.Contains(t, logger.Warnings[1], "score board not saved")
	assert.Contains(t, logger.Warnings[1], "score board not loaded", "should report the load failure as the reason the save was skipped")
	assert.Nil(t, store.Saved)
}

func TestIntegrationRunnerRunPrintsWarningAndSkipsSaveWhenSavingScoreFails(t *testing.T) {
	events := fake.NewInputEventSource()
	events.Ch <- input.QuitEvent{}

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
	assert.Contains(t, logger.Warnings[0], "score board not saved")
	assert.Contains(t, logger.Warnings[0], "disk full")
}

func TestIntegrationRunnerRunWillSaveScoreWhenScoreBoardWasNotFound(t *testing.T) {
	events := fake.NewInputEventSource()
	events.Ch <- input.QuitEvent{}

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
	assert.Contains(t, logger.Warnings[0], "score board not loaded")
}

func TestIntegrationRunnerRunWillQuitOnQuitEvent(t *testing.T) {
	events := fake.NewInputEventSource()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Ch <- input.QuitEvent{}
	}()

	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 100, Name: "target", Rss: 1024}}},
		&fake.Store{},
		events,
		&fake.Logger{},
	)

	start := time.Now()
	err := r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0})

	require.NoError(t, err)
	assert.Less(t, time.Since(start), time.Second, "Run should have quit shortly after the QuitEvent, not run indefinitely")
}

func TestIntegrationRunnerRunWillQuitWhenTimeLimitExpires(t *testing.T) {
	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 102, Name: "target", Rss: 1024}}},
		&fake.Store{},
		fake.NewInputEventSource(),
		&fake.Logger{},
	)

	start := time.Now()
	require.NoError(t, r.Run(inbound.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 1}))
	assert.LessOrEqual(t, time.Since(start), 3*time.Second, "game took too long to exit on time limit")
}

// TestIntegrationRunnerRunSignalGoroutineDoesNotAccumulate verifies the signal goroutine
// started inside runLoop exits when Run returns, preventing goroutine leaks.
func TestIntegrationRunnerRunSignalGoroutineDoesNotAccumulate(t *testing.T) {
	runGame := func(pid int) {
		events := fake.NewInputEventSource()
		go func() {
			time.Sleep(50 * time.Millisecond)
			events.Ch <- input.QuitEvent{}
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
