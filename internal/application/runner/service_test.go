package runner

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestServiceRunReturnsErrorWhenRunningAsRootWithoutOverride(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0)}}
	r := newTestServiceWithProcess(&testutil.FakeProcessManager{OwnUIDValue: 0}, []string{"proc"}, opts)

	err := r.Run()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
	assert.ErrorContains(t, err, "refusing to run as root")
}

func TestServiceRunReturnsErrorWhenSpeedTooLow(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(movement.MinSpeed - 0.1)}}
	err := newTestService([]string{"proc"}, opts).Run()
	assertFatal(t, err)
}

func TestServiceRunReturnsErrorWhenSpeedTooHigh(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(movement.MaxSpeed + 0.1)}}
	err := newTestService([]string{"proc"}, opts).Run()
	assertFatal(t, err)
}

func TestServiceRunReturnsErrorWhenTimeLimitIsNegative(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0), TimeLimit: testutil.Pointer(-1)}}
	err := newTestService([]string{"proc"}, opts).Run()
	assertFatal(t, err)
}

func TestServiceRunReturnsErrorWhenNoPatternsAreProvided(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0)}}
	err := newTestService(nil, opts).Run()
	assertFatal(t, err)
	assert.ErrorContains(t, err, "at least one search pattern is required")
}

func TestServiceRunReturnsErrorWhenPatternTooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0)}}
		err := newTestService([]string{p}, opts).Run()
		assertFatal(t, err)
	}
}

func TestServiceRunReturnsErrorWhenProcessDiscoveryFails(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: testutil.Pointer(2.0)}}
	r := newTestServiceWithProcess(&testutil.FakeProcessManager{DiscoverErr: errors.New("ps failed"), OwnUIDValue: 1000}, []string{"proc"}, opts)

	err := r.Run()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessNotFound, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}

func newTestService(patterns []string, opts config.Options) *Service {
	return newTestServiceWithProcess(&testutil.FakeProcessManager{OwnUIDValue: 1000}, patterns, opts)
}

func newTestServiceWithProcess(proc *testutil.FakeProcessManager, patterns []string, opts config.Options) *Service {
	configSvc := config.NewService(&testutil.FakeConfigStore{LoadErr: outbound.NotFoundError{}}, opts)
	processSvc := process.NewService(proc, &testutil.FakeProcessReporter{}, patterns)
	scoreSvc := score.NewService(&testutil.FakeStore{}, &testutil.FakeScoreReporter{})
	gameSvc := game.NewService(processSvc, &testutil.FakeRenderer{}, testutil.NewFakeInputEventProvider())
	errHandler := apperror.NewHandler(&testutil.FakeLogger{})
	return NewService(configSvc, processSvc, scoreSvc, gameSvc, errHandler)
}

func assertFatal(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}
