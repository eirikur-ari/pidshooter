package application

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func newTestRunner() *Runner {
	return NewRunner(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource(), &fake.Logger{})
}

func assertFatal(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}

func TestRunnerRunSpeedTooLow(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Patterns: []string{"proc"}, Speed: movement.MinSpeed - 0.1})
	assertFatal(t, err)
}

func TestRunnerRunSpeedTooHigh(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Patterns: []string{"proc"}, Speed: movement.MaxSpeed + 0.1})
	assertFatal(t, err)
}

func TestRunnerRunNegativeTimeLimit(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Patterns: []string{"proc"}, Speed: 2.0, TimeLimit: -1})
	assertFatal(t, err)
}

func TestRunnerRunNoPatterns(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Speed: 2.0})
	assertFatal(t, err)
	assert.ErrorContains(t, err, "at least one search pattern is required")
}

func TestRunnerRunPatternTooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		err := newTestRunner().Run(inbound.Config{Patterns: []string{p}, Speed: 2.0})
		assertFatal(t, err)
	}
}

func TestRunnerRunPatternExactMinLength(t *testing.T) {
	min := strings.Repeat("a", process.MinPatternLength)
	logger := &fake.Logger{}
	r := NewRunner(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource(), logger)

	err := r.Run(inbound.Config{Patterns: []string{min}, Speed: 2.0})

	assertFatal(t, err)
	require.Len(t, logger.Errors, 1)
	assert.Contains(t, logger.Errors[0], "no processes found")
}

func TestRunnerRunPatternTooLong(t *testing.T) {
	long := strings.Repeat("a", process.MaxPatternLength+1)
	err := newTestRunner().Run(inbound.Config{Patterns: []string{long}, Speed: 2.0})
	assertFatal(t, err)
}

func TestRunnerRunProcessDiscoveryFailure(t *testing.T) {
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

	r.logKillFailures([]string{
		"could not kill proc (PID 123): boom",
		"could not kill other (PID 456): bam",
	})

	require.Len(t, logger.Warnings, 2)
	assert.Contains(t, logger.Warnings[0], "could not kill proc (PID 123): boom")
	assert.Contains(t, logger.Warnings[1], "could not kill other (PID 456): bam")
}
