package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
)

func TestProgram_Run_PrintsHelpToStdout(t *testing.T) {
	// Given
	fixture := newProgramFixture()

	// When
	err := fixture.program.Run(nil)

	// Then
	require.NoError(t, err)
	assert.Contains(t, fixture.out.String(), "Process ID Shooter")
	assert.Empty(t, fixture.errOut.String())
	assert.Empty(t, fixture.logger.Errored)
}

func TestProgram_Run_LogsArgumentErrorOnceAndPrintsUsageToStderr(t *testing.T) {
	// Given
	fixture := newProgramFixture()

	// When
	err := fixture.program.Run([]string{"play", "proc", "--unknown"})

	// Then
	require.Error(t, err)
	assert.ErrorAs(t, err, &ArgumentError{})
	assert.Empty(t, fixture.out.String())
	assert.Contains(t, fixture.errOut.String(), "Usage:")
	assert.Equal(t, []string{err.Error()}, fixture.logger.Errored)
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
		})
	}
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
	assert.Empty(t, fixture.logger.Warned)
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
	}
}
