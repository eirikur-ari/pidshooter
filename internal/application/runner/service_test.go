package runner

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
	"github.com/eirikur-ari/pidshooter/internal/testutil/helper"
)

func TestServiceRunReturnsErrorWhenSpeedTooLow(t *testing.T) {
	err := newTestService().Run(inbound.RunRequest{Patterns: []string{"proc"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(movement.MinSpeed - 0.1)}}})
	assertFatal(t, err)
}

func TestServiceRunReturnsErrorWhenSpeedTooHigh(t *testing.T) {
	err := newTestService().Run(inbound.RunRequest{Patterns: []string{"proc"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(movement.MaxSpeed + 0.1)}}})
	assertFatal(t, err)
}

func TestServiceRunReturnsErrorWhenTimeLimitIsNegative(t *testing.T) {
	err := newTestService().Run(inbound.RunRequest{Patterns: []string{"proc"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(-1)}}})
	assertFatal(t, err)
}

func TestServiceRunReturnsErrorWhenNoPatternsAreProvided(t *testing.T) {
	err := newTestService().Run(inbound.RunRequest{Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0)}}})
	assertFatal(t, err)
	assert.ErrorContains(t, err, "at least one search pattern is required")
}

func TestServiceRunReturnsErrorWhenPatternTooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		err := newTestService().Run(inbound.RunRequest{Patterns: []string{p}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0)}}})
		assertFatal(t, err)
	}
}

func TestServiceRunReturnsErrorWhenProcessDiscoveryFails(t *testing.T) {
	r := newTestServiceWithProcess(&fake.Process{DiscoverErr: errors.New("ps failed")})

	err := r.Run(inbound.RunRequest{Patterns: []string{"proc"}, Config: config.Request{Game: config.GameRequest{Speed: helper.Ptr(2.0)}}})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessDiscoveryFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}

func newTestService() *Service {
	return newTestServiceWithProcess(&fake.Process{})
}

func newTestServiceWithProcess(proc *fake.Process) *Service {
	configSvc := config.NewService(&fake.ConfigStore{LoadErr: outbound.NotFoundError{}})
	processSvc := process.NewService(proc, &fake.ProcessReporter{})
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
