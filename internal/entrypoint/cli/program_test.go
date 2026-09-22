package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func TestRunNoArgsPrintsUsageAndReturnsNil(t *testing.T) {
	assert.NoError(t, newTestProgram(&fake.Runner{}).Run([]string{}))
}

func TestRunHelpFlagPrintsUsageAndReturnsNil(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			assert.NoError(t, newTestProgram(&fake.Runner{}).Run([]string{flag}))
		})
	}
}

func TestRunHelpWinsRegardlessOfPositionOrOtherErrors(t *testing.T) {
	tests := [][]string{
		{"--help", "--unknown"},
		{"--unknown", "--help"},
		{"proc", "--speed", "--help"},
		{"--help", "--speed"},
	}
	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			assert.NoError(t, newTestProgram(&fake.Runner{}).Run(args))
		})
	}
}

func TestRunMalformedFlagReturnsArgumentError(t *testing.T) {
	tests := [][]string{
		{"proc", "--help=true"}, // not an exact -h/--help match; falls through as an unrecognized flag
		{"proc", "--help=false"},
		{"proc", "--unknown"},
		{"proc", "--speed"},     // missing its value
		{"proc", "--speed=abc"}, // wrong type
		{"proc", "--time=abc"},  // wrong type
	}
	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			err := newTestProgram(&fake.Runner{}).Run(args)
			require.Error(t, err)
			assert.True(t, errors.As(err, &ArgumentError{}))
		})
	}
}

func TestRunBasicPattern(t *testing.T) {
	runner := &fake.Runner{}
	require.NoError(t, newTestProgram(runner).Run([]string{"firefox"}))
	require.Len(t, runner.Cfg.Patterns, 1)
	assert.Equal(t, "firefox", runner.Cfg.Patterns[0])
	assert.False(t, runner.Cfg.ConfirmMode)
	assert.Equal(t, 2.0, runner.Cfg.Speed)
	assert.Equal(t, 30, runner.Cfg.TimeLimit)
}

func TestRunMultiplePatterns(t *testing.T) {
	runner := &fake.Runner{}
	require.NoError(t, newTestProgram(runner).Run([]string{"chrome", "firefox", "node"}))
	require.Len(t, runner.Cfg.Patterns, 3)
	for i, want := range []string{"chrome", "firefox", "node"} {
		assert.Equal(t, want, runner.Cfg.Patterns[i])
	}
}

func TestRunConfirmFlag(t *testing.T) {
	runner := &fake.Runner{}
	require.NoError(t, newTestProgram(runner).Run([]string{"sleep", "--confirm"}))
	assert.True(t, runner.Cfg.ConfirmMode)
}

func TestRunIncludeRootFlag(t *testing.T) {
	runner := &fake.Runner{}
	require.NoError(t, newTestProgram(runner).Run([]string{"sleep", "--include-root"}))
	assert.True(t, runner.Cfg.IncludeRoot)
}

func TestRunIncludeRootDefaultsToFalse(t *testing.T) {
	runner := &fake.Runner{}
	require.NoError(t, newTestProgram(runner).Run([]string{"sleep"}))
	assert.False(t, runner.Cfg.IncludeRoot)
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
			assertFlagResult(t, tt.arg, tt.wantErr, tt.want, func(cfg config.Config) float64 { return cfg.Speed })
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
			assertFlagResult(t, tt.arg, tt.wantErr, tt.want, func(cfg config.Config) int { return cfg.TimeLimit })
		})
	}
}

func TestRunPatternAfterFlag(t *testing.T) {
	runner := &fake.Runner{}
	require.NoError(t, newTestProgram(runner).Run([]string{"chrome", "--confirm", "firefox"}))
	assert.Equal(t, []string{"chrome", "firefox"}, runner.Cfg.Patterns)
	assert.True(t, runner.Cfg.ConfirmMode)
}

func TestRunFlagsSurroundingPatterns(t *testing.T) {
	runner := &fake.Runner{}
	require.NoError(t, newTestProgram(runner).Run([]string{"chrome", "--speed", "3.5", "firefox", "node", "--confirm"}))
	assert.Equal(t, []string{"chrome", "firefox", "node"}, runner.Cfg.Patterns)
	assert.Equal(t, 3.5, runner.Cfg.Speed)
	assert.True(t, runner.Cfg.ConfirmMode)
}

func TestRunFlagsWithoutPatternsIsRejected(t *testing.T) {
	assert.Error(t, newTestProgram(&fake.Runner{}).Run([]string{"--confirm"}))
}

func TestRunNoArgsDoesNotConstructRunner(t *testing.T) {
	creator := &fake.RunnerCreator{}
	require.NoError(t, NewProgram(creator).Run([]string{}))
	assert.Zero(t, creator.Calls)
}

func TestRunHelpDoesNotConstructRunner(t *testing.T) {
	creator := &fake.RunnerCreator{}
	require.NoError(t, NewProgram(creator).Run([]string{"--help"}))
	assert.Zero(t, creator.Calls)
}

func TestRunInvalidFlagsDoesNotConstructRunner(t *testing.T) {
	creator := &fake.RunnerCreator{}
	require.Error(t, NewProgram(creator).Run([]string{"proc", "--unknown"}))
	assert.Zero(t, creator.Calls)
}

func TestRunInvalidConfigDoesNotConstructRunner(t *testing.T) {
	creator := &fake.RunnerCreator{}
	require.Error(t, NewProgram(creator).Run([]string{"--confirm"}))
	assert.Zero(t, creator.Calls)
}

func TestRunReturnsErrorWhenRunnerCreatorFails(t *testing.T) {
	creator := &fake.RunnerCreator{Err: errors.New("boom")}

	err := NewProgram(creator).Run([]string{"proc"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
	assert.Equal(t, 1, creator.Calls)
}

func TestRunNoArgsPrintsUsageToStdout(t *testing.T) {
	program, out, errOut := newCapturingTestProgram(&fake.Runner{})

	require.NoError(t, program.Run([]string{}))

	assert.Contains(t, out.String(), "Process ID Shooter")
	assert.Empty(t, errOut.String())
}

func TestRunHelpPrintsUsageToStdout(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			program, out, errOut := newCapturingTestProgram(&fake.Runner{})

			require.NoError(t, program.Run([]string{flag}))

			assert.Contains(t, out.String(), "Process ID Shooter")
			assert.Empty(t, errOut.String())
		})
	}
}

func TestRunUnknownFlagPrintsUsageToStderr(t *testing.T) {
	program, out, errOut := newCapturingTestProgram(&fake.Runner{})

	require.Error(t, program.Run([]string{"proc", "--unknown"}))

	assert.Empty(t, out.String())
	assert.Contains(t, errOut.String(), "Process ID Shooter")
}

func TestRunInvalidConfigPrintsUsageToStderr(t *testing.T) {
	program, out, errOut := newCapturingTestProgram(&fake.Runner{})

	require.Error(t, program.Run([]string{"--confirm"}))

	assert.Empty(t, out.String())
	assert.Contains(t, errOut.String(), "Process ID Shooter")
}

func TestRunReturnsErrorOnFatal(t *testing.T) {
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	runner := &fake.Runner{Err: fatal}

	err := newTestProgram(runner).Run([]string{"proc"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "game session failed")
}

func newTestProgram(runner inbound.Runner) *Program {
	return NewProgram(&fake.RunnerCreator{Runner: runner})
}

func newCapturingTestProgram(runner inbound.Runner) (program *Program, out, errOut *bytes.Buffer) {
	out, errOut = &bytes.Buffer{}, &bytes.Buffer{}
	program = &Program{creator: &fake.RunnerCreator{Runner: runner}, out: out, errOut: errOut}
	return program, out, errOut
}

func assertFlagResult[T any](t *testing.T, arg string, wantErr bool, want T, get func(config.Config) T) {
	t.Helper()
	runner := &fake.Runner{}
	err := newTestProgram(runner).Run([]string{"proc", arg})
	if wantErr {
		assert.Error(t, err)
		return
	}
	require.NoError(t, err)
	assert.Equal(t, want, get(runner.Cfg))
}
