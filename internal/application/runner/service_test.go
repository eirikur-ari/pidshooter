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
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
	"github.com/eirikur-ari/pidshooter/internal/testutil/helper"
)

func TestServiceRunReturnsErrorWhenRunningAsRootWithoutOverride(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: helper.Pointer(2.0)}}
	r := newTestServiceWithProcess(&fake.Process{OwnUIDValue: 0}, []string{"proc"}, opts)

	err := r.Run()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
	assert.ErrorContains(t, err, "refusing to run as root")
}

func TestServiceRunReturnsErrorWhenSpeedTooLow(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: helper.Pointer(movement.MinSpeed - 0.1)}}
	err := newTestService([]string{"proc"}, opts).Run()
	assertFatal(t, err)
}

func TestServiceRunReturnsErrorWhenSpeedTooHigh(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: helper.Pointer(movement.MaxSpeed + 0.1)}}
	err := newTestService([]string{"proc"}, opts).Run()
	assertFatal(t, err)
}

func TestServiceRunReturnsErrorWhenTimeLimitIsNegative(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: helper.Pointer(2.0), TimeLimit: helper.Pointer(-1)}}
	err := newTestService([]string{"proc"}, opts).Run()
	assertFatal(t, err)
}

func TestServiceRunReturnsErrorWhenNoPatternsAreProvided(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: helper.Pointer(2.0)}}
	err := newTestService(nil, opts).Run()
	assertFatal(t, err)
	assert.ErrorContains(t, err, "at least one search pattern is required")
}

func TestServiceRunReturnsErrorWhenPatternTooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		opts := config.Options{Game: config.GameOptions{Speed: helper.Pointer(2.0)}}
		err := newTestService([]string{p}, opts).Run()
		assertFatal(t, err)
	}
}

func TestServiceRunReturnsErrorWhenProcessDiscoveryFails(t *testing.T) {
	opts := config.Options{Game: config.GameOptions{Speed: helper.Pointer(2.0)}}
	r := newTestServiceWithProcess(&fake.Process{DiscoverErr: errors.New("ps failed"), OwnUIDValue: 1000}, []string{"proc"}, opts)

	err := r.Run()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessDiscoveryFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}

func newTestService(patterns []string, opts config.Options) *Service {
	return newTestServiceWithProcess(&fake.Process{OwnUIDValue: 1000}, patterns, opts)
}

func newTestServiceWithProcess(proc *fake.Process, patterns []string, opts config.Options) *Service {
	configSvc := config.NewService(&fake.ConfigStore{LoadErr: outbound.NotFoundError{}}, opts)
	processSvc := process.NewService(proc, &fake.ProcessReporter{}, patterns)
	scoreSvc := score.NewService(&fake.Store{}, &fake.ScoreReporter{})
	gameSvc := game.NewService(processSvc, &fake.Renderer{}, fake.NewInputEventProvider())
	errHandler := apperror.NewHandler(&fake.Logger{})
	return NewService(configSvc, processSvc, scoreSvc, gameSvc, errHandler)
}

func assertFatal(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}
