package application

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func TestRunnerRunReturnsErrorWhenSpeedTooLow(t *testing.T) {
	err := newTestRunner().Run(config.Config{Patterns: []string{"proc"}, Speed: movement.MinSpeed - 0.1})
	assertFatal(t, err)
}

func TestRunnerRunReturnsErrorWhenSpeedTooHigh(t *testing.T) {
	err := newTestRunner().Run(config.Config{Patterns: []string{"proc"}, Speed: movement.MaxSpeed + 0.1})
	assertFatal(t, err)
}

func TestRunnerRunReturnsErrorWhenTimeLimitIsNegative(t *testing.T) {
	err := newTestRunner().Run(config.Config{Patterns: []string{"proc"}, Speed: 2.0, TimeLimit: -1})
	assertFatal(t, err)
}

func TestRunnerRunReturnsErrorWhenNoPatternsAreProvided(t *testing.T) {
	err := newTestRunner().Run(config.Config{Speed: 2.0})
	assertFatal(t, err)
	assert.ErrorContains(t, err, "at least one search pattern is required")
}

func TestRunnerRunReturnsErrorWhenPatternTooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		err := newTestRunner().Run(config.Config{Patterns: []string{p}, Speed: 2.0})
		assertFatal(t, err)
	}
}

func TestRunnerRunReturnsErrorWhenProcessDiscoveryFails(t *testing.T) {
	r := newTestRunnerWithProcess(&fake.Process{DiscoverErr: errors.New("ps failed")})

	err := r.Run(config.Config{Patterns: []string{"proc"}, Speed: 2.0})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessDiscoveryFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}

func newTestRunner() *Runner {
	return newTestRunnerWithProcess(&fake.Process{})
}

func newTestRunnerWithProcess(proc *fake.Process) *Runner {
	processSvc := process.NewService(proc, &fake.ProcessReporter{})
	scoreSvc := score.NewService(&fake.Store{}, &fake.ScoreReporter{})
	gameSvc := game.NewService(processSvc, &fake.Renderer{}, fake.NewInputEventProvider())
	return NewRunner(processSvc, scoreSvc, gameSvc)
}

func assertFatal(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}
