package application

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func TestRunnerRunReturnsErrorWhenSpeedTooLow(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Patterns: []string{"proc"}, Speed: movement.MinSpeed - 0.1})
	assertFatal(t, err)
}

func TestRunnerRunReturnsErrorWhenSpeedTooHigh(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Patterns: []string{"proc"}, Speed: movement.MaxSpeed + 0.1})
	assertFatal(t, err)
}

func TestRunnerRunReturnsErrorWhenTimeLimitIsNegative(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Patterns: []string{"proc"}, Speed: 2.0, TimeLimit: -1})
	assertFatal(t, err)
}

func TestRunnerRunReturnsErrorWhenNoPatternsAreProvided(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Speed: 2.0})
	assertFatal(t, err)
	assert.ErrorContains(t, err, "at least one search pattern is required")
}

func TestRunnerRunReturnsErrorWhenPatternTooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		err := newTestRunner().Run(inbound.Config{Patterns: []string{p}, Speed: 2.0})
		assertFatal(t, err)
	}
}

func TestRunnerRunReturnsErrorWhenPatternExactMinLength(t *testing.T) {
	minPatternLength := strings.Repeat("a", process.MinPatternLength)
	logger := &fake.Logger{}
	r := NewRunner(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource(), logger)

	err := r.Run(inbound.Config{Patterns: []string{minPatternLength}, Speed: 2.0})

	assertFatal(t, err)
	require.Len(t, logger.Errors, 1)
	assert.Contains(t, logger.Errors[0], "no processes found")
}

func TestRunnerRunConfigValidationReturnsErrorWhenPatternTooLong(t *testing.T) {
	long := strings.Repeat("a", process.MaxPatternLength+1)
	err := newTestRunner().Run(inbound.Config{Patterns: []string{long}, Speed: 2.0})
	assertFatal(t, err)
}

func TestRunnerRunReturnsErrorWhenProcessDiscoveryFails(t *testing.T) {
	logger := &fake.Logger{}
	r := NewRunner(&fake.Process{ListErr: errors.New("ps failed")}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource(), logger)

	err := r.Run(inbound.Config{Patterns: []string{"proc"}, Speed: 2.0})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessDiscoveryFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	require.Len(t, logger.Errors, 1)
	assert.Contains(t, logger.Errors[0], "ps failed")
}

func TestRunnerLogKillFailuresLogsEachAsWarning(t *testing.T) {
	logger := &fake.Logger{}
	r := NewRunner(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource(), logger)

	r.logKillFailures([]game.KillFailure{
		{Target: "proc", PID: 123, Err: errors.New("boom")},
		{Target: "other", PID: 456, Err: errors.New("bam")},
	})

	require.Len(t, logger.Warnings, 2)
	assert.Equal(t, "could not kill proc (PID 123): boom", logger.Warnings[0])
	assert.Equal(t, "could not kill other (PID 456): bam", logger.Warnings[1])
}

func newTestRunner() *Runner {
	return NewRunner(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource(), &fake.Logger{})
}

func assertFatal(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}
