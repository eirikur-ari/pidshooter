package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestProgram_Run_PrintsUsageToStdoutWithoutCreatingRunner(t *testing.T) {
	tests := [][]string{
		{},
		{"--help"},
		{"-h"},
		{"--help", "--unknown"},
		{"--unknown", "--help"},
		{"proc", "--speed", "--help"},
		{"--help", "--speed"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			// Given
			creator := &FakeRunnerCreator{}
			program, out, errOut := newProgramFixture(creator)

			// When
			err := program.Run(args)

			// Then
			require.NoError(t, err)
			assert.Equal(t, usageText, out.String())
			assert.Empty(t, errOut.String())
			assert.Zero(t, creator.Calls)
		})
	}
}

func TestProgram_Run_ReturnsArgumentErrorAndPrintsUsageToStderrWhenArgumentsAreMalformed(t *testing.T) {
	tests := [][]string{
		{"proc", "--help=true"}, // not an exact -h/--help match; falls through as an unrecognized flag
		{"proc", "--help=false"},
		{"proc", "--unknown"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			// Given
			creator := &FakeRunnerCreator{}
			program, out, errOut := newProgramFixture(creator)

			// When
			err := program.Run(args)

			// Then
			require.Error(t, err)
			assert.True(t, errors.As(err, &ArgumentError{}))
			assert.Empty(t, out.String())
			assert.Contains(t, errOut.String(), usageText)
			assert.Zero(t, creator.Calls)
		})
	}
}

func TestProgram_Run_PassesPatternsAndOptionsToRunnerCreator(t *testing.T) {
	// Given
	runner := &FakeRunner{}
	creator := &FakeRunnerCreator{Runner: runner}
	program, _, _ := newProgramFixture(creator)

	// When
	err := program.Run([]string{"chrome", "--confirm", "firefox", "--speed=3.5"})

	// Then
	require.NoError(t, err)
	assert.Equal(t, 1, creator.Calls)
	assert.Equal(t, []string{"chrome", "firefox"}, creator.Patterns)
	assert.Equal(t, testutil.Pointer(true), creator.Options.Game.ConfirmMode)
	assert.Equal(t, testutil.Pointer(3.5), creator.Options.Game.Speed)
	assert.Nil(t, creator.Options.Game.TimeLimit)
	assert.Equal(t, 1, runner.Calls)
}

func TestProgram_Run_ReturnsErrorWhenRunnerCreatorFails(t *testing.T) {
	// Given
	creationError := errors.New("boom")
	creator := &FakeRunnerCreator{Err: creationError}
	program, out, errOut := newProgramFixture(creator)

	// When
	err := program.Run([]string{"proc"})

	// Then
	require.ErrorIs(t, err, creationError)
	assert.Equal(t, 1, creator.Calls)
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
}

func TestProgram_Run_ReclassifiesInvalidConfigAsArgumentError(t *testing.T) {
	// Given
	invalid := apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", errors.New("boom"))
	runner := &FakeRunner{Err: invalid}
	creator := &FakeRunnerCreator{Runner: runner}
	program, out, errOut := newProgramFixture(creator)

	// When
	err := program.Run([]string{"proc"})

	// Then
	require.Error(t, err)
	assert.True(t, errors.As(err, &ArgumentError{}))
	assert.Empty(t, out.String())
	assert.Contains(t, errOut.String(), usageText)
	assert.Equal(t, 1, runner.Calls)
}

func TestProgram_Run_ReturnsOtherFatalErrorsUnchanged(t *testing.T) {
	// Given
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	runner := &FakeRunner{Err: fatal}
	creator := &FakeRunnerCreator{Runner: runner}
	program, out, errOut := newProgramFixture(creator)

	// When
	err := program.Run([]string{"proc"})

	// Then
	require.ErrorIs(t, err, fatal)
	assert.False(t, errors.As(err, &ArgumentError{}))
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
	assert.Equal(t, 1, runner.Calls)
}

func TestProgram_Run_AbsorbsWarningFromRunner(t *testing.T) {
	// Given
	warning := apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, "could not kill target", errors.New("boom"))
	runner := &FakeRunner{Err: warning}
	creator := &FakeRunnerCreator{Runner: runner}
	program, out, errOut := newProgramFixture(creator)

	// When
	err := program.Run([]string{"proc"})

	// Then
	assert.NoError(t, err)
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
	assert.Equal(t, 1, runner.Calls)
}
