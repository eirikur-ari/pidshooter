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
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestIntegrationServiceRunIsSuccessful(t *testing.T) {
	events := testutil.NewFakeInputEventProvider()
	events.Ch <- input.QuitEvent{}

	store := &testutil.FakeStore{}
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0), TimeLimit: testutil.Pointer(0)}}
	r := newService(
		&testutil.FakeProcessManager{Infos: []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024, UID: 1000}}, OwnUIDValue: 1000},
		store,
		events,
		[]string{"target"},
		opts,
	)

	err := r.Run()
	require.NoError(t, err)
	require.NotNil(t, store.Saved)
	assert.Empty(t, store.Saved.Scores, "quitting immediately with zero kills must not be recorded to the score board")
}

func TestIntegrationServiceRunAsRootWithOverrideReachesProcessDiscovery(t *testing.T) {
	events := testutil.NewFakeInputEventProvider()
	events.Ch <- input.QuitEvent{}

	opts := config.Options{
		Game:    config.GameOptions{Speed: testutil.Pointer(2.0), TimeLimit: testutil.Pointer(0)},
		Process: config.ProcessOptions{AllowRoot: testutil.Pointer(true)},
	}
	r := newService(
		&testutil.FakeProcessManager{Infos: []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024}}, OwnUIDValue: 0},
		&testutil.FakeStore{},
		events,
		[]string{"target"},
		opts,
	)

	err := r.Run()

	require.NoError(t, err, "AllowRoot must let a root session reach process discovery")
}

func TestIntegrationServiceIncludeRootFalseExcludesRootOwnedProcess(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0), TimeLimit: testutil.Pointer(0)}}
	r := newService(
		&testutil.FakeProcessManager{
			Infos:       []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024, UID: 0}},
			OwnUIDValue: 1000,
		},
		&testutil.FakeStore{},
		testutil.NewFakeInputEventProvider(),
		[]string{"target"},
		opts,
	)

	err := r.Run()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessNotFound, appErr.Code)
}

func TestIntegrationServiceIncludeRootTrueIncludesRootOwnedProcess(t *testing.T) {
	events := testutil.NewFakeInputEventProvider()
	events.Ch <- input.QuitEvent{}

	opts := config.Options{
		Game:    config.GameOptions{Speed: testutil.Pointer(2.0), TimeLimit: testutil.Pointer(0)},
		Process: config.ProcessOptions{IncludeRoot: testutil.Pointer(true)},
	}
	r := newService(
		&testutil.FakeProcessManager{
			Infos:       []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024, UID: 0}},
			OwnUIDValue: 1000,
		},
		&testutil.FakeStore{},
		events,
		[]string{"target"},
		opts,
	)

	err := r.Run()

	require.NoError(t, err)
}

func TestIntegrationServicePersistedIncludeRootIncludesRootOwnedProcess(t *testing.T) {
	events := testutil.NewFakeInputEventProvider()
	events.Ch <- input.QuitEvent{}

	configStore := &testutil.FakeConfigStore{Config: outbound.Config{Process: outbound.ProcessConfig{IncludeRoot: testutil.Pointer(true)}}}
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0), TimeLimit: testutil.Pointer(0)}}
	r := newServiceWithConfigStore(
		&testutil.FakeProcessManager{
			Infos:       []outbound.ProcessInfo{{PID: 200, Name: "target", Rss: 1024, UID: 0}},
			OwnUIDValue: 1000,
		},
		&testutil.FakeStore{},
		events,
		&testutil.FakeRenderer{},
		[]string{"target"},
		opts,
		configFakes{store: configStore, logger: &testutil.FakeLogger{}},
	)

	err := r.Run()

	require.NoError(t, err, "a persisted include_root: true, with no request override, must reach process discovery")
}

func TestIntegrationServicePersistedTimeLimitQuitsGameplay(t *testing.T) {
	configStore := &testutil.FakeConfigStore{Config: outbound.Config{Game: outbound.GameConfig{TimeLimit: testutil.Pointer(1)}}}
	r := newServiceWithConfigStore(
		&testutil.FakeProcessManager{Infos: []outbound.ProcessInfo{{PID: 102, Name: "target", Rss: 1024, UID: 1000}}, OwnUIDValue: 1000},
		&testutil.FakeStore{},
		testutil.NewFakeInputEventProvider(),
		&testutil.FakeRenderer{},
		[]string{"target"},
		config.Options{},
		configFakes{store: configStore, logger: &testutil.FakeLogger{}},
	)

	start := time.Now()
	err := r.Run()

	require.NoError(t, err)
	assert.LessOrEqual(t, time.Since(start), 3*time.Second, "a persisted time_limit, with no request override, must reach gameplay and end the session")
}

func TestIntegrationServiceFallsBackAndWarnsOnUnreadableConfigStore(t *testing.T) {
	events := testutil.NewFakeInputEventProvider()
	events.Ch <- input.QuitEvent{}

	configStore := &testutil.FakeConfigStore{LoadErr: outbound.CorruptedDataError{Message: "not valid yaml"}}
	logger := &testutil.FakeLogger{}
	r := newServiceWithConfigStore(
		&testutil.FakeProcessManager{Infos: []outbound.ProcessInfo{{PID: 205, Name: "target", Rss: 1024, UID: 1000}}, OwnUIDValue: 1000},
		&testutil.FakeStore{},
		events,
		&testutil.FakeRenderer{},
		[]string{"target"},
		config.Options{},
		configFakes{store: configStore, logger: logger},
	)

	err := r.Run()

	require.NoError(t, err, "an unreadable config store must fall back to domain defaults, not fail the run")
	assert.NotEmpty(t, logger.Warned, "the unreadable config store must still be reported as a warning")
}

func TestIntegrationServiceRunReturnsErrorWhenRendererInitFails(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0), TimeLimit: testutil.Pointer(0)}}
	r := newServiceWithRenderer(
		&testutil.FakeProcessManager{Infos: []outbound.ProcessInfo{{PID: 204, Name: "target", Rss: 1024, UID: 1000}}, OwnUIDValue: 1000},
		&testutil.FakeStore{},
		testutil.NewFakeInputEventProvider(),
		&testutil.FakeRenderer{InitErr: errors.New("terminal not available")},
		[]string{"target"},
		opts,
	)

	err := r.Run()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeGameFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}

func TestIntegrationServiceRunSkipsSaveWhenLoadingScoreBoardFails(t *testing.T) {
	events := testutil.NewFakeInputEventProvider()
	events.Ch <- input.QuitEvent{}

	store := &testutil.FakeStore{LoadErr: errors.New("json: invalid character")}
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0), TimeLimit: testutil.Pointer(0)}}
	r := newService(
		&testutil.FakeProcessManager{Infos: []outbound.ProcessInfo{{PID: 201, Name: "target", Rss: 1024, UID: 1000}}, OwnUIDValue: 1000},
		store,
		events,
		[]string{"target"},
		opts,
	)

	err := r.Run()

	require.NoError(t, err)
	assert.Nil(t, store.Saved)
}

func TestIntegrationServiceRunDoesNotFailWhenSavingScoreFails(t *testing.T) {
	events := testutil.NewFakeInputEventProvider()
	events.Ch <- input.QuitEvent{}

	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0), TimeLimit: testutil.Pointer(0)}}
	r := newService(
		&testutil.FakeProcessManager{Infos: []outbound.ProcessInfo{{PID: 202, Name: "target", Rss: 1024, UID: 1000}}, OwnUIDValue: 1000},
		&testutil.FakeStore{SaveErr: errors.New("disk full")},
		events,
		[]string{"target"},
		opts,
	)

	err := r.Run()

	require.NoError(t, err)
}

func TestIntegrationServiceRunWillSaveScoreWhenScoreBoardWasNotFound(t *testing.T) {
	events := testutil.NewFakeInputEventProvider()
	events.Ch <- input.QuitEvent{}

	store := &testutil.FakeStore{LoadErr: outbound.NotFoundError{}}
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0), TimeLimit: testutil.Pointer(0)}}
	r := newService(
		&testutil.FakeProcessManager{Infos: []outbound.ProcessInfo{{PID: 203, Name: "target", Rss: 1024, UID: 1000}}, OwnUIDValue: 1000},
		store,
		events,
		[]string{"target"},
		opts,
	)

	err := r.Run()

	require.NoError(t, err)
	require.NotNil(t, store.Saved, "a fresh (never-persisted) board should still be saved")
}

func TestIntegrationServiceRunWillQuitOnQuitEvent(t *testing.T) {
	events := testutil.NewFakeInputEventProvider()
	go func() {
		time.Sleep(50 * time.Millisecond)
		events.Ch <- input.QuitEvent{}
	}()

	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0)}}
	r := newService(
		&testutil.FakeProcessManager{Infos: []outbound.ProcessInfo{{PID: 100, Name: "target", Rss: 1024, UID: 1000}}, OwnUIDValue: 1000},
		&testutil.FakeStore{},
		events,
		[]string{"target"},
		opts,
	)

	start := time.Now()
	err := r.Run()

	require.NoError(t, err)
	assert.Less(t, time.Since(start), time.Second, "Run should have quit shortly after the QuitEvent, not run indefinitely")
}

func TestIntegrationServiceRunWillQuitWhenTimeLimitExpires(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0), TimeLimit: testutil.Pointer(1)}}
	r := newService(
		&testutil.FakeProcessManager{Infos: []outbound.ProcessInfo{{PID: 102, Name: "target", Rss: 1024, UID: 1000}}, OwnUIDValue: 1000},
		&testutil.FakeStore{},
		testutil.NewFakeInputEventProvider(),
		[]string{"target"},
		opts,
	)

	start := time.Now()
	require.NoError(t, r.Run())
	assert.LessOrEqual(t, time.Since(start), 3*time.Second, "game took too long to exit on time limit")
}

// TestIntegrationServiceRunSignalGoroutineDoesNotAccumulate verifies the signal goroutine
// started inside runLoop exits when Run returns, preventing goroutine leaks.
func TestIntegrationServiceRunSignalGoroutineDoesNotAccumulate(t *testing.T) {
	runGame := func(pid int) {
		events := testutil.NewFakeInputEventProvider()
		go func() {
			time.Sleep(50 * time.Millisecond)
			events.Ch <- input.QuitEvent{}
		}()
		opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0)}}
		r := newService(
			&testutil.FakeProcessManager{Infos: []outbound.ProcessInfo{{PID: pid, Name: "target", Rss: 1024, UID: 1000}}, OwnUIDValue: 1000},
			&testutil.FakeStore{},
			events,
			[]string{"target"},
			opts,
		)
		if err := r.Run(); err != nil {
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

func newService(proc *testutil.FakeProcessManager, store *testutil.FakeStore, events *testutil.FakeInputEventProvider, patterns []string, opts config.Options) *Service {
	return newServiceWithRenderer(proc, store, events, &testutil.FakeRenderer{}, patterns, opts)
}

func newServiceWithRenderer(proc *testutil.FakeProcessManager, store *testutil.FakeStore, events *testutil.FakeInputEventProvider, renderer *testutil.FakeRenderer, patterns []string, opts config.Options) *Service {
	return newServiceWithConfigStore(proc, store, events, renderer, patterns, opts, configFakes{store: &testutil.FakeConfigStore{LoadErr: outbound.NotFoundError{}}, logger: &testutil.FakeLogger{}})
}

// configFakes bundles the config-store fakes newServiceWithConfigStore
// wires in, kept together since callers always override both or neither.
type configFakes struct {
	store  *testutil.FakeConfigStore
	logger *testutil.FakeLogger
}

func newServiceWithConfigStore(proc *testutil.FakeProcessManager, store *testutil.FakeStore, events *testutil.FakeInputEventProvider, renderer *testutil.FakeRenderer, patterns []string, opts config.Options, cfg configFakes) *Service {
	configSvc := config.NewService(cfg.store, opts)
	processSvc := process.NewService(proc, &testutil.FakeProcessReporter{}, patterns)
	scoreSvc := score.NewService(store, &testutil.FakeScoreReporter{})
	gameSvc := game.NewService(processSvc, renderer, events)
	errHandler := apperror.NewHandler(cfg.logger)
	return NewService(configSvc, processSvc, scoreSvc, gameSvc, errHandler)
}
