package cli

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func TestRunNoArgsPrintsUsageAndReturnsNil(t *testing.T) {
	assert.NoError(t, newTestCLI(&fake.Runner{}).Run([]string{}))
}

func TestRunHelpFlagPrintsUsageAndReturnsNil(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			assert.NoError(t, newTestCLI(&fake.Runner{}).Run([]string{flag}))
		})
	}
}

func TestRunBasicPattern(t *testing.T) {
	runner := &fake.Runner{}
	require.NoError(t, newTestCLI(runner).Run([]string{"firefox"}))
	require.Len(t, runner.Cfg.Patterns, 1)
	assert.Equal(t, "firefox", runner.Cfg.Patterns[0])
	assert.False(t, runner.Cfg.ConfirmMode)
	assert.Equal(t, 2.0, runner.Cfg.Speed)
	assert.Equal(t, 30, runner.Cfg.TimeLimit)
}

func TestRunMultiplePatterns(t *testing.T) {
	runner := &fake.Runner{}
	require.NoError(t, newTestCLI(runner).Run([]string{"chrome", "firefox", "node"}))
	require.Len(t, runner.Cfg.Patterns, 3)
	for i, want := range []string{"chrome", "firefox", "node"} {
		assert.Equal(t, want, runner.Cfg.Patterns[i])
	}
}

func TestRunConfirmFlag(t *testing.T) {
	runner := &fake.Runner{}
	require.NoError(t, newTestCLI(runner).Run([]string{"sleep", "--confirm"}))
	assert.True(t, runner.Cfg.ConfirmMode)
}

func TestRunSpeedFlag(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    float64
		wantErr bool
	}{
		{"valid speed", "--speed=3.5", 3.5, false},
		{"min speed", "--speed=0.5", 0.5, false},
		{"max speed", "--speed=5.0", 5.0, false},
		{"invalid", "--speed=abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertFlagResult(t, tt.arg, tt.wantErr, tt.want, func(cfg inbound.Config) float64 { return cfg.Speed })
		})
	}
}

func TestRunTimeFlag(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    int
		wantErr bool
	}{
		{"valid time", "--time=60", 60, false},
		{"no limit", "--time=0", 0, false},
		{"invalid", "--time=abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertFlagResult(t, tt.arg, tt.wantErr, tt.want, func(cfg inbound.Config) int { return cfg.TimeLimit })
		})
	}
}

func TestRunUnknownFlag(t *testing.T) {
	assert.Error(t, newTestCLI(&fake.Runner{}).Run([]string{"proc", "--unknown"}))
}

func TestRunNoArgsWithFlagsForwardsToService(t *testing.T) {
	runner := &fake.Runner{}
	require.NoError(t, newTestCLI(runner).Run([]string{"--confirm"}))
	assert.Empty(t, runner.Cfg.Patterns)
	assert.True(t, runner.Cfg.ConfirmMode)
}

func TestRunLogsUnrecognizedError(t *testing.T) {
	runner := &fake.Runner{Err: errors.New("boom")}
	logger := &fake.Logger{}
	c := newTestCLI(runner)
	c.logger = logger

	err := c.Run([]string{"proc"})

	require.Error(t, err)
	require.Len(t, logger.Errors, 1)
	assert.Contains(t, logger.Errors[0], "boom")
}

func TestRunDoesNotLogAppError(t *testing.T) {
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	runner := &fake.Runner{Err: fatal}
	logger := &fake.Logger{}
	c := newTestCLI(runner)
	c.logger = logger

	err := c.Run([]string{"proc"})

	require.Error(t, err)
	assert.Empty(t, logger.Errors, "an *apperror.Error was already logged by the application layer")
}

func TestRunDoesNotLogWrappedAppError(t *testing.T) {
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	runner := &fake.Runner{Err: fmt.Errorf("during Run: %w", fatal)}
	logger := &fake.Logger{}
	c := newTestCLI(runner)
	c.logger = logger

	err := c.Run([]string{"proc"})

	require.Error(t, err)
	assert.Empty(t, logger.Errors, "a wrapped *apperror.Error was already logged by the application layer")
}

func TestRunReturnsErrorOnFatal(t *testing.T) {
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	runner := &fake.Runner{Err: fatal}

	err := newTestCLI(runner).Run([]string{"proc"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "game session failed")
}

func newTestCLI(runner inbound.Runner) *CLI {
	c := NewCLI(runner, &fake.Logger{})
	c.out = io.Discard
	c.errOut = io.Discard
	return c
}

// assertFlagResult runs the CLI with arg and asserts either a parse error, or
// that get extracts want from the resulting Config.
func assertFlagResult[T any](t *testing.T, arg string, wantErr bool, want T, get func(inbound.Config) T) {
	t.Helper()
	runner := &fake.Runner{}
	err := newTestCLI(runner).Run([]string{"proc", arg})
	if wantErr {
		assert.Error(t, err)
		return
	}
	require.NoError(t, err)
	assert.Equal(t, want, get(runner.Cfg))
}
