//go:build integration

package runner

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
	"github.com/eirikur-ari/pidshooter/internal/testutil/helper"
)

func TestIntegrationServiceRunIsSuccessful(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	store := &fake.Store{}
	r := newService(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024}}},
		store,
		events,
	)

	err := r.Run(inbound.RunRequest{Patterns: []string{"target"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(0)}}})
	require.NoError(t, err)
	require.NotNil(t, store.Saved)
	assert.Empty(t, store.Saved.Scores, "quitting immediately with zero kills must not be recorded to the score board")
}

func TestIntegrationServiceIncludeRootFalseExcludesRootOwnedProcess(t *testing.T) {
	r := newService(
		&fake.Process{
			Infos:       []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024, UID: 0}},
			OwnUIDValue: 1000,
		},
		&fake.Store{},
		fake.NewInputEventProvider(),
	)

	err := r.Run(inbound.RunRequest{Patterns: []string{"target"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(0)}}})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessNotFound, appErr.Code)
}

func TestIntegrationServiceIncludeRootTrueIncludesRootOwnedProcess(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	r := newService(
		&fake.Process{
			Infos:       []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024, UID: 0}},
			OwnUIDValue: 1000,
		},
		&fake.Store{},
		events,
	)

	err := r.Run(inbound.RunRequest{
		Patterns: []string{"target"},
		Config: config.Request{
			Game:    config.GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(0)},
			Process: config.ProcessRequest{IncludeRoot: helper.Ptr(true)},
		},
	})

	require.NoError(t, err)
}

func TestIntegrationServicePersistedIncludeRootIncludesRootOwnedProcess(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	configStore := &fake.ConfigStore{Result: outbound.ConfigStoreResult{Process: outbound.ProcessConfig{IncludeRoot: helper.Ptr(true)}}}
	r := newServiceWithConfigStore(
		&fake.Process{
			Infos:       []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024, UID: 0}},
			OwnUIDValue: 1000,
		},
		&fake.Store{},
		events,
		&fake.Renderer{},
		configStore,
		&fake.Logger{},
	)

	err := r.Run(inbound.RunRequest{Patterns: []string{"target"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(0)}}})

	require.NoError(t, err, "a persisted include_root: true, with no request override, must reach process discovery")
}

func TestIntegrationServicePersistedTimeLimitQuitsGameplay(t *testing.T) {
	configStore := &fake.ConfigStore{Result: outbound.ConfigStoreResult{Game: outbound.GameConfig{TimeLimit: helper.Ptr(1)}}}
	r := newServiceWithConfigStore(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 102, Name: "target", Rss: 1024}}},
		&fake.Store{},
		fake.NewInputEventProvider(),
		&fake.Renderer{},
		configStore,
		&fake.Logger{},
	)

	start := time.Now()
	err := r.Run(inbound.RunRequest{Patterns: []string{"target"}})

	require.NoError(t, err)
	assert.LessOrEqual(t, time.Since(start), 3*time.Second, "a persisted time_limit, with no request override, must reach gameplay and end the session")
}

func TestIntegrationServiceFallsBackAndWarnsOnUnreadableConfigStore(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	configStore := &fake.ConfigStore{LoadErr: outbound.CorruptedDataError{Message: "not valid yaml"}}
	logger := &fake.Logger{}
	r := newServiceWithConfigStore(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 205, Name: "target", Rss: 1024}}},
		&fake.Store{},
		events,
		&fake.Renderer{},
		configStore,
		logger,
	)

	err := r.Run(inbound.RunRequest{Patterns: []string{"target"}})

	require.NoError(t, err, "an unreadable config store must fall back to domain defaults, not fail the run")
	assert.NotEmpty(t, logger.Warned, "the unreadable config store must still be reported as a warning")
}

func TestIntegrationServiceRunReturnsErrorWhenRendererInitFails(t *testing.T) {
	r := newServiceWithRenderer(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 204, Name: "target", Rss: 1024}}},
		&fake.Store{},
		fake.NewInputEventProvider(),
		&fake.Renderer{InitErr: errors.New("terminal not available")},
	)

	err := r.Run(inbound.RunRequest{Patterns: []string{"target"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(0)}}})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeGameFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}

func TestIntegrationServiceRunSkipsSaveWhenLoadingScoreBoardFails(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	store := &fake.Store{LoadErr: errors.New("json: invalid character")}
	r := newService(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 201, Name: "target", Rss: 1024}}},
		store,
		events,
	)

	err := r.Run(inbound.RunRequest{Patterns: []string{"target"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(0)}}})

	require.NoError(t, err)
	assert.Nil(t, store.Saved)
}

func TestIntegrationServiceRunDoesNotFailWhenSavingScoreFails(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	r := newService(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 202, Name: "target", Rss: 1024}}},
		&fake.Store{SaveErr: errors.New("disk full")},
		events,
	)

	err := r.Run(inbound.RunRequest{Patterns: []string{"target"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(0)}}})

	require.NoError(t, err)
}

func TestIntegrationServiceRunWillSaveScoreWhenScoreBoardWasNotFound(t *testing.T) {
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}

	store := &fake.Store{LoadErr: outbound.NotFoundError{}}
	r := newService(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 203, Name: "target", Rss: 1024}}},
		store,
		events,
	)

	err := r.Run(inbound.RunRequest{Patterns: []string{"target"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(0)}}})

	require.NoError(t, err)
	require.NotNil(t, store.Saved, "a fresh (never-persisted) board should still be saved")
}

func TestIntegrationServiceRunWillQuitOnQuitEvent(t *testing.T) {
	events := fake.NewInputEventProvider()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Ch <- input.QuitEvent{}
	}()

	r := newService(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 100, Name: "target", Rss: 1024}}},
		&fake.Store{},
		events,
	)

	start := time.Now()
	err := r.Run(inbound.RunRequest{Patterns: []string{"target"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0)}}})

	require.NoError(t, err)
	assert.Less(t, time.Since(start), time.Second, "Run should have quit shortly after the QuitEvent, not run indefinitely")
}

func TestIntegrationServiceRunWillQuitWhenTimeLimitExpires(t *testing.T) {
	r := newService(
		&fake.Process{Infos: []outbound.ProcessInfo{{PID: 102, Name: "target", Rss: 1024}}},
		&fake.Store{},
		fake.NewInputEventProvider(),
	)

	start := time.Now()
	require.NoError(t, r.Run(inbound.RunRequest{Patterns: []string{"target"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(1)}}}))
	assert.LessOrEqual(t, time.Since(start), 3*time.Second, "game took too long to exit on time limit")
}

// TestIntegrationServiceRunSignalGoroutineDoesNotAccumulate verifies the signal goroutine
// started inside runLoop exits when Run returns, preventing goroutine leaks.
func TestIntegrationServiceRunSignalGoroutineDoesNotAccumulate(t *testing.T) {
	runGame := func(pid int) {
		events := fake.NewInputEventProvider()
		go func() {
			time.Sleep(50 * time.Millisecond)
			events.Ch <- input.QuitEvent{}
		}()
		r := newService(
			&fake.Process{Infos: []outbound.ProcessInfo{{PID: pid, Name: "target", Rss: 1024}}},
			&fake.Store{},
			events,
		)
		if err := r.Run(inbound.RunRequest{Patterns: []string{"target"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0)}}}); err != nil {
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

func newService(proc *fake.Process, store *fake.Store, events *fake.InputEventProvider) *Service {
	return newServiceWithRenderer(proc, store, events, &fake.Renderer{})
}

func newServiceWithRenderer(proc *fake.Process, store *fake.Store, events *fake.InputEventProvider, renderer *fake.Renderer) *Service {
	return newServiceWithConfigStore(proc, store, events, renderer, &fake.ConfigStore{LoadErr: outbound.NotFoundError{}}, &fake.Logger{})
}

func newServiceWithConfigStore(proc *fake.Process, store *fake.Store, events *fake.InputEventProvider, renderer *fake.Renderer, configStore *fake.ConfigStore, logger *fake.Logger) *Service {
	configSvc := config.NewService(configStore)
	processSvc := process.NewService(proc, &fake.ProcessReporter{})
	scoreSvc := score.NewService(store, &fake.ScoreReporter{})
	gameSvc := game.NewService(processSvc, renderer, events)
	errHandler := apperror.NewHandler(logger)
	return NewService(configSvc, processSvc, scoreSvc, gameSvc, errHandler)
}
