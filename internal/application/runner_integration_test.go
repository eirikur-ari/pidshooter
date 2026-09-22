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
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func newRunner(proc *fake.Process, store *fake.Store, events *fake.InputEventProvider) *Runner {
	return NewRunner(proc, store, &fake.ScoreReporter{}, &fake.Renderer{}, events)
}

func TestIntegrationRunnerRunIsSuccessful(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	store := &fake.Store{}
	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024}}},
		store,
		events,
	)

	err := r.Run(config.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})
	require.NoError(t, err)
	require.NotNil(t, store.Saved)
	assert.Empty(t, store.Saved.Scores, "quitting immediately with zero kills must not be recorded to the score board")
}

func TestIntegrationRunnerIncludeRootFalseExcludesRootOwnedProcess(t *testing.T) {
	r := newRunner(
		&fake.Process{
			Infos:       []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024, UID: 0}},
			OwnUIDValue: 1000,
		},
		&fake.Store{},
		fake.NewInputEventProvider(),
	)

	err := r.Run(config.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessNotFound, appErr.Code)
}

func TestIntegrationRunnerIncludeRootTrueIncludesRootOwnedProcess(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	r := newRunner(
		&fake.Process{
			Infos:       []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024, UID: 0}},
			OwnUIDValue: 1000,
		},
		&fake.Store{},
		events,
	)

	err := r.Run(config.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0, IncludeRoot: true})

	require.NoError(t, err)
}

func TestIntegrationRunnerRunReturnsErrorWhenRendererInitFails(t *testing.T) {
	r := NewRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 204, Name: "target", Rss: 1024}}},
		&fake.Store{},
		&fake.ScoreReporter{},
		&fake.Renderer{InitErr: errors.New("terminal not available")},
		fake.NewInputEventProvider(),
	)

	err := r.Run(config.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeGameFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}

func TestIntegrationRunnerRunSkipsSaveWhenLoadingScoreBoardFails(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	store := &fake.Store{LoadErr: errors.New("json: invalid character")}
	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 201, Name: "target", Rss: 1024}}},
		store,
		events,
	)

	err := r.Run(config.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})

	require.NoError(t, err)
	assert.Nil(t, store.Saved)
}

func TestIntegrationRunnerRunDoesNotFailWhenSavingScoreFails(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 202, Name: "target", Rss: 1024}}},
		&fake.Store{SaveErr: errors.New("disk full")},
		events,
	)

	err := r.Run(config.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})

	require.NoError(t, err)
}

func TestIntegrationRunnerRunWillSaveScoreWhenScoreBoardWasNotFound(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	store := &fake.Store{LoadErr: outbound.NotFoundError{}}
	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 203, Name: "target", Rss: 1024}}},
		store,
		events,
	)

	err := r.Run(config.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})

	require.NoError(t, err)
	require.NotNil(t, store.Saved, "a fresh (never-persisted) board should still be saved")
}

func TestIntegrationRunnerRunWillQuitOnQuitEvent(t *testing.T) {
	events := fake.NewInputEventProvider()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Ch <- input.QuitEvent{}
	}()

	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 100, Name: "target", Rss: 1024}}},
		&fake.Store{},
		events,
	)

	start := time.Now()
	err := r.Run(config.Config{Patterns: []string{"target"}, Speed: 2.0})

	require.NoError(t, err)
	assert.Less(t, time.Since(start), time.Second, "Run should have quit shortly after the QuitEvent, not run indefinitely")
}

func TestIntegrationRunnerRunWillQuitWhenTimeLimitExpires(t *testing.T) {
	r := newRunner(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 102, Name: "target", Rss: 1024}}},
		&fake.Store{},
		fake.NewInputEventProvider(),
	)

	start := time.Now()
	require.NoError(t, r.Run(config.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 1}))
	assert.LessOrEqual(t, time.Since(start), 3*time.Second, "game took too long to exit on time limit")
}

// TestIntegrationRunnerRunSignalGoroutineDoesNotAccumulate verifies the signal goroutine
// started inside runLoop exits when Run returns, preventing goroutine leaks.
func TestIntegrationRunnerRunSignalGoroutineDoesNotAccumulate(t *testing.T) {
	runGame := func(pid int) {
		events := fake.NewInputEventProvider()
		go func() {
			time.Sleep(50 * time.Millisecond)
			events.Ch <- input.QuitEvent{}
		}()
		r := newRunner(
			&fake.Process{Infos: []outbound.ProcessInfo{{PID: pid, Name: "target", Rss: 1024}}},
			&fake.Store{},
			events,
		)
		if err := r.Run(config.Config{Patterns: []string{"target"}, Speed: 2.0}); err != nil {
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
