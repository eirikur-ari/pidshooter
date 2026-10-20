package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
)

func TestProgram_Run_PrintsHelpToStdoutWithoutRunningAnything(t *testing.T) {
	tests := newHelpTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			fixture := newProgramFixture()

			// When
			err := fixture.program.Run(test.args)

			// Then
			require.NoError(t, err)
			assert.Contains(t, fixture.out.String(), test.expectedOutput)
			assert.Empty(t, fixture.errOut.String())
			assert.Empty(t, fixture.logger.Errored)
			assertNoRunnerRan(t, fixture)
		})
	}
}

func TestProgram_Run_ReturnsArgumentErrorLogsItOnceAndPrintsUsageToStderrWhenArgumentsAreMalformed(t *testing.T) {
	tests := [][]string{
		{"play", "proc", "--unknown"},
		{"list", "proc", "--unknown"},
		{"kill", "proc", "--unknown"},
		{"list", "proc", "--speed=3"}, // a flag of play, not of list
		{"kill", "proc", "--confirm"}, // a flag of play, not of kill
		{"play", "proc", "--speed"},   // missing its value
		{"play", "proc", "--speed=abc"},
		{"play", "proc", "--time=abc"},
		{"chrome"},
		{"chrome", "play"},
		{"Play", "chrome"},
		{"play"},
		{"list"},
		{"kill"},
		{"--include-root", "play"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			// Given
			fixture := newProgramFixture()

			// When
			err := fixture.program.Run(args)

			// Then
			require.Error(t, err)
			assert.True(t, errors.As(err, &ArgumentError{}))
			assert.Empty(t, fixture.out.String())
			assert.Contains(t, fixture.errOut.String(), "Usage:")
			assert.Equal(t, []string{err.Error()}, fixture.logger.Errored)
			assertNoRunnerRan(t, fixture)
		})
	}
}

func TestProgram_Run_RunsOnlyPlayRunnerForTheCommand(t *testing.T) {
	// Given
	fixture := newProgramFixture()
	returnFromEveryRunner(fixture, nil)

	// When
	err := fixture.program.Run([]string{"play", "proc"})

	// Then
	require.NoError(t, err)
	fixture.play.AssertNumberOfCalls(t, "Run", 1)
	fixture.list.AssertNotCalled(t, "Run", mock.Anything)
	fixture.kill.AssertNotCalled(t, "Run", mock.Anything)
	assert.Empty(t, fixture.out.String())
	assert.Empty(t, fixture.errOut.String())
	assert.Empty(t, fixture.logger.Errored)
}

func TestProgram_Run_RunsOnlyListRunnerForTheCommand(t *testing.T) {
	// Given
	fixture := newProgramFixture()
	returnFromEveryRunner(fixture, nil)

	// When
	err := fixture.program.Run([]string{"list", "proc"})

	// Then
	require.NoError(t, err)
	fixture.list.AssertNumberOfCalls(t, "Run", 1)
	fixture.play.AssertNotCalled(t, "Run", mock.Anything)
	fixture.kill.AssertNotCalled(t, "Run", mock.Anything)
	assert.Empty(t, fixture.out.String())
	assert.Empty(t, fixture.errOut.String())
	assert.Empty(t, fixture.logger.Errored)
}

func TestProgram_Run_RunsOnlyKillRunnerForTheCommand(t *testing.T) {
	// Given
	fixture := newProgramFixture()
	returnFromEveryRunner(fixture, nil)

	// When
	err := fixture.program.Run([]string{"kill", "proc"})

	// Then
	require.NoError(t, err)
	fixture.kill.AssertNumberOfCalls(t, "Run", 1)
	fixture.play.AssertNotCalled(t, "Run", mock.Anything)
	fixture.list.AssertNotCalled(t, "Run", mock.Anything)
	assert.Empty(t, fixture.out.String())
	assert.Empty(t, fixture.errOut.String())
	assert.Empty(t, fixture.logger.Errored)
}

func TestProgram_Run_ReturnsErrorOfRunnerUnchangedWithoutLoggingIt(t *testing.T) {
	// Given
	failure := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "session failed", errors.New("boom"))
	fixture := newProgramFixture()
	returnFromEveryRunner(fixture, failure)

	// When
	err := fixture.program.Run([]string{"play", "proc"})

	// Then
	assert.Same(t, failure, err)
	assert.False(t, errors.As(err, &ArgumentError{}))
	assert.Empty(t, fixture.out.String())
	assert.Empty(t, fixture.errOut.String())
	assert.Empty(t, fixture.logger.Errored)
	assert.Empty(t, fixture.logger.Warned)
}

func TestProgram_Run_ReclassifiesInvalidConfigAsArgumentErrorWithoutLoggingIt(t *testing.T) {
	// Given
	invalid := apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", errors.New("boom"))
	fixture := newProgramFixture()
	returnFromEveryRunner(fixture, invalid)

	// When
	err := fixture.program.Run([]string{"play", "proc"})

	// Then
	require.ErrorIs(t, err, invalid)
	assert.True(t, errors.As(err, &ArgumentError{}))
	assert.Empty(t, fixture.out.String())
	assert.Contains(t, fixture.errOut.String(), "pidshooter play <pattern>")
	assert.Empty(t, fixture.logger.Errored)
}

func TestProgram_Run_PrintsUsageOfTheCommandWhenItsArgumentsAreMalformed(t *testing.T) {
	tests := newCommandUsageTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			fixture := newProgramFixture()

			// When
			err := fixture.program.Run(test.args)

			// Then
			require.Error(t, err)
			assert.Contains(t, fixture.errOut.String(), test.expectedUsage)
			assertNoRunnerRan(t, fixture)
		})
	}
}

func newHelpTestCase() []struct {
	name           string
	args           []string
	expectedOutput string
} {
	return []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{"no arguments", nil, "Process ID Shooter"},
		{"help flag", []string{"--help"}, "Process ID Shooter"},
		{"short help flag", []string{"-h"}, "Process ID Shooter"},
		{"play help", []string{"play", "--help"}, "pidshooter play"},
		{"list help", []string{"list", "-h"}, "pidshooter list"},
		{"kill help", []string{"kill", "--help"}, "pidshooter kill"},
		{"help after pattern", []string{"play", "proc", "--help"}, "pidshooter play"},
	}
}

func newCommandUsageTestCase() []struct {
	name          string
	args          []string
	expectedUsage string
} {
	return []struct {
		name          string
		args          []string
		expectedUsage string
	}{
		{"play without patterns", []string{"play"}, "pidshooter play <pattern>"},
		{"list without patterns", []string{"list"}, "pidshooter list <pattern>"},
		{"kill without patterns", []string{"kill"}, "pidshooter kill <pattern>"},
		{"play with unknown flag", []string{"play", "proc", "--unknown"}, "pidshooter play <pattern>"},
	}
}
