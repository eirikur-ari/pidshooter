package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
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
			creator := &MockRunnerCreator{}
			program, out, errOut := newProgramFixture(creator)

			// When
			err := program.Run(args)

			// Then
			require.NoError(t, err)
			assert.Equal(t, usageText, out.String())
			assert.Empty(t, errOut.String())
			creator.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
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
			creator := &MockRunnerCreator{}
			program, out, errOut := newProgramFixture(creator)

			// When
			err := program.Run(args)

			// Then
			require.Error(t, err)
			assert.True(t, errors.As(err, &ArgumentError{}))
			assert.Empty(t, out.String())
			assert.Contains(t, errOut.String(), usageText)
			creator.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		})
	}
}

func TestProgram_Run_PassesPatternsAndOptionsToRunnerCreator(t *testing.T) {
	// Given
	expectedOptions := config.Options{
		Game: config.GameOptions{
			ConfirmMode: new(true),
			Speed:       new(3.5),
		},
	}
	runner := &MockRunner{}
	runner.On("Run").Return(nil)
	creator := &MockRunnerCreator{}
	creator.On("Create", []string{"chrome", "firefox"}, expectedOptions).Return(runner, nil)
	program, _, _ := newProgramFixture(creator)

	// When
	err := program.Run([]string{"chrome", "--confirm", "firefox", "--speed=3.5"})

	// Then
	require.NoError(t, err)
	creator.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestProgram_Run_ReturnsErrorWhenRunnerCreatorFails(t *testing.T) {
	// Given
	creationError := errors.New("boom")
	creator := &MockRunnerCreator{}
	creator.On("Create", mock.Anything, mock.Anything).Return(nil, creationError)
	program, out, errOut := newProgramFixture(creator)

	// When
	err := program.Run([]string{"proc"})

	// Then
	require.ErrorIs(t, err, creationError)
	creator.AssertNumberOfCalls(t, "Create", 1)
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
}

func TestProgram_Run_ReclassifiesInvalidConfigAsArgumentError(t *testing.T) {
	// Given
	invalid := apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", errors.New("boom"))
	creator, runner := runnerCreatorMockSetup(invalid)
	program, out, errOut := newProgramFixture(creator)

	// When
	err := program.Run([]string{"proc"})

	// Then
	require.Error(t, err)
	assert.True(t, errors.As(err, &ArgumentError{}))
	assert.Empty(t, out.String())
	assert.Contains(t, errOut.String(), usageText)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestProgram_Run_ReturnsOtherFatalErrorsUnchanged(t *testing.T) {
	// Given
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	creator, runner := runnerCreatorMockSetup(fatal)
	program, out, errOut := newProgramFixture(creator)

	// When
	err := program.Run([]string{"proc"})

	// Then
	require.ErrorIs(t, err, fatal)
	assert.False(t, errors.As(err, &ArgumentError{}))
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestProgram_Run_AbsorbsWarningFromRunner(t *testing.T) {
	// Given
	warning := apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, "could not kill target", errors.New("boom"))
	creator, runner := runnerCreatorMockSetup(warning)
	program, out, errOut := newProgramFixture(creator)

	// When
	err := program.Run([]string{"proc"})

	// Then
	assert.NoError(t, err)
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
	runner.AssertNumberOfCalls(t, "Run", 1)
}
