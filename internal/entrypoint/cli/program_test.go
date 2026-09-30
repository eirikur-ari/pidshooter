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
	program, creator := newTestProgramWithCreator(runner)
	require.NoError(t, program.Run([]string{"firefox"}))
	require.Len(t, creator.Patterns, 1)
	assert.Equal(t, "firefox", creator.Patterns[0])
	assert.Nil(t, creator.Options.Game.ConfirmMode, "an unpassed flag must stay nil, leaving it to the resolved defaults")
	assert.Nil(t, creator.Options.Game.Speed)
	assert.Nil(t, creator.Options.Game.TimeLimit)
}

func TestRunMultiplePatterns(t *testing.T) {
	runner := &fake.Runner{}
	program, creator := newTestProgramWithCreator(runner)
	require.NoError(t, program.Run([]string{"chrome", "firefox", "node"}))
	require.Len(t, creator.Patterns, 3)
	for i, want := range []string{"chrome", "firefox", "node"} {
		assert.Equal(t, want, creator.Patterns[i])
	}
}

func TestRunConfirmFlag(t *testing.T) {
	runner := &fake.Runner{}
	program, creator := newTestProgramWithCreator(runner)
	require.NoError(t, program.Run([]string{"sleep", "--confirm"}))
	require.NotNil(t, creator.Options.Game.ConfirmMode)
	assert.True(t, *creator.Options.Game.ConfirmMode)
}

func TestRunIncludeRootFlag(t *testing.T) {
	runner := &fake.Runner{}
	program, creator := newTestProgramWithCreator(runner)
	require.NoError(t, program.Run([]string{"sleep", "--include-root"}))
	require.NotNil(t, creator.Options.Process.IncludeRoot)
	assert.True(t, *creator.Options.Process.IncludeRoot)
}

func TestRunIncludeRootUnsetWhenFlagAbsent(t *testing.T) {
	runner := &fake.Runner{}
	program, creator := newTestProgramWithCreator(runner)
	require.NoError(t, program.Run([]string{"sleep"}))
	assert.Nil(t, creator.Options.Process.IncludeRoot)
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
			assertFlagResult(t, tt.arg, tt.wantErr, tt.want, func(opts config.Options) *float64 { return opts.Game.Speed })
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
			assertFlagResult(t, tt.arg, tt.wantErr, tt.want, func(opts config.Options) *int { return opts.Game.TimeLimit })
		})
	}
}

func TestRunPatternAfterFlag(t *testing.T) {
	runner := &fake.Runner{}
	program, creator := newTestProgramWithCreator(runner)
	require.NoError(t, program.Run([]string{"chrome", "--confirm", "firefox"}))
	assert.Equal(t, []string{"chrome", "firefox"}, creator.Patterns)
	require.NotNil(t, creator.Options.Game.ConfirmMode)
	assert.True(t, *creator.Options.Game.ConfirmMode)
}

func TestRunFlagsSurroundingPatterns(t *testing.T) {
	runner := &fake.Runner{}
	program, creator := newTestProgramWithCreator(runner)
	require.NoError(t, program.Run([]string{"chrome", "--speed", "3.5", "firefox", "node", "--confirm"}))
	assert.Equal(t, []string{"chrome", "firefox", "node"}, creator.Patterns)
	require.NotNil(t, creator.Options.Game.Speed)
	assert.Equal(t, 3.5, *creator.Options.Game.Speed)
	require.NotNil(t, creator.Options.Game.ConfirmMode)
	assert.True(t, *creator.Options.Game.ConfirmMode)
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
	invalid := apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", errors.New("boom"))
	program, out, errOut := newCapturingTestProgram(&fake.Runner{Err: invalid})

	require.Error(t, program.Run([]string{"proc"}))

	assert.Empty(t, out.String())
	assert.Contains(t, errOut.String(), "Process ID Shooter")
}

func TestRunReclassifiesInvalidConfigAsArgumentError(t *testing.T) {
	invalid := apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", errors.New("boom"))
	runner := &fake.Runner{Err: invalid}

	err := newTestProgram(runner).Run([]string{"proc"})

	require.Error(t, err)
	assert.True(t, errors.As(err, &ArgumentError{}))
}

func TestRunDoesNotReclassifyOtherFatalErrorsAsArgumentError(t *testing.T) {
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	runner := &fake.Runner{Err: fatal}

	err := newTestProgram(runner).Run([]string{"proc"})

	require.Error(t, err)
	assert.False(t, errors.As(err, &ArgumentError{}))
}

func TestRunReturnsErrorOnFatal(t *testing.T) {
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	runner := &fake.Runner{Err: fatal}

	err := newTestProgram(runner).Run([]string{"proc"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "game session failed")
}

func TestRunAbsorbsWarningFromRunner(t *testing.T) {
	warning := apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, "could not kill target", errors.New("boom"))
	runner := &fake.Runner{Err: warning}

	err := newTestProgram(runner).Run([]string{"proc"})

	assert.NoError(t, err)
}

func newTestProgram(runner inbound.Runner) *Program {
	return NewProgram(&fake.RunnerCreator{Runner: runner})
}

func newTestProgramWithCreator(runner inbound.Runner) (*Program, *fake.RunnerCreator) {
	creator := &fake.RunnerCreator{Runner: runner}
	return NewProgram(creator), creator
}

func newCapturingTestProgram(runner inbound.Runner) (program *Program, out, errOut *bytes.Buffer) {
	out, errOut = &bytes.Buffer{}, &bytes.Buffer{}
	creator := &fake.RunnerCreator{Runner: runner}
	program = &Program{creator: creator, errHandler: creator.ErrHandler(), out: out, errOut: errOut}
	return program, out, errOut
}

func assertFlagResult[T any](t *testing.T, arg string, wantErr bool, want T, get func(config.Options) *T) {
	t.Helper()
	runner := &fake.Runner{}
	program, creator := newTestProgramWithCreator(runner)
	err := program.Run([]string{"proc", arg})
	if wantErr {
		assert.Error(t, err)
		return
	}
	require.NoError(t, err)
	got := get(creator.Options)
	require.NotNil(t, got)
	assert.Equal(t, want, *got)
}
