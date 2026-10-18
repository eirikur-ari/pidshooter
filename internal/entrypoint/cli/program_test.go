package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestProgram_Run_PrintsUsageToStdoutWithoutCreatingRunner(t *testing.T) {
	tests := [][]string{
		{},
		{"--help"},
		{"-h"},
		{"--help", "--unknown"},
		{"--unknown", "--help"},
		{"play", "proc", "--speed", "--help"},
		{"--help", "--speed"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			// Given
			runner := newRunnerMock()
			program, out, errOut := newProgramFixture(runner)

			// When
			err := program.Run(args)

			// Then
			require.NoError(t, err)
			assert.Equal(t, usageText, out.String())
			assert.Empty(t, errOut.String())
			runner.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

func TestProgram_Run_ReturnsArgumentErrorAndPrintsUsageToStderrWhenArgumentsAreMalformed(t *testing.T) {
	tests := [][]string{
		{"play", "proc", "--help=true"}, // not an exact -h/--help match; falls through as an unrecognized flag
		{"play", "proc", "--help=false"},
		{"play", "proc", "--unknown"},
		{"chrome"},
		{"chrome", "play"},
		{"--time=15", "play", "chrome"},
		{"Play", "chrome"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			// Given
			runner := newRunnerMock()
			program, out, errOut := newProgramFixture(runner)

			// When
			err := program.Run(args)

			// Then
			require.Error(t, err)
			assert.True(t, errors.As(err, &ArgumentError{}))
			assert.Empty(t, out.String())
			assert.Contains(t, errOut.String(), usageText)
			runner.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

func TestProgram_Run_PassesPlayRequestToPlayRunner(t *testing.T) {
	// Given
	expectedRequest := inbound.PlayRequest{
		Patterns: []string{"chrome", "firefox"},
		Game: inbound.GameOptions{
			ConfirmMode: new(true),
			Speed:       new(3.5),
		},
	}
	runner := newRunnerMock()
	runner.On("Play", expectedRequest).Return(nil)
	args := []string{"play", "chrome", "--confirm", "firefox", "--speed=3.5"}
	program, _, _ := newProgramFixture(runner)

	// When
	err := program.Run(args)

	// Then
	require.NoError(t, err)
	runner.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Play", 1)
}

func TestProgram_Run_ReclassifiesInvalidConfigAsArgumentError(t *testing.T) {
	// Given
	invalid := apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", errors.New("boom"))
	runner := runnerMockSetup(invalid)
	args := []string{"play", "proc"}
	program, out, errOut := newProgramFixture(runner)

	// When
	err := program.Run(args)

	// Then
	require.Error(t, err)
	assert.True(t, errors.As(err, &ArgumentError{}))
	assert.Empty(t, out.String())
	assert.Contains(t, errOut.String(), usageText)
	runner.AssertNumberOfCalls(t, "Play", 1)
}

func TestProgram_Run_ReturnsOtherFatalErrorsUnchanged(t *testing.T) {
	// Given
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	runner := runnerMockSetup(fatal)
	args := []string{"play", "proc"}
	program, out, errOut := newProgramFixture(runner)

	// When
	err := program.Run(args)

	// Then
	require.ErrorIs(t, err, fatal)
	assert.False(t, errors.As(err, &ArgumentError{}))
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
	runner.AssertNumberOfCalls(t, "Play", 1)
}

func TestProgram_Run_AbsorbsWarningFromRunner(t *testing.T) {
	// Given
	warning := apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, "could not kill target", errors.New("boom"))
	runner := &MockRunner{}
	runner.On("Play", mock.Anything).Return(warning)
	runner.On("HandleError", warning).Return(nil)
	args := []string{"play", "proc"}
	program, out, errOut := newProgramFixture(runner)

	// When
	err := program.Run(args)

	// Then
	assert.NoError(t, err)
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
	runner.AssertNumberOfCalls(t, "Play", 1)
}
