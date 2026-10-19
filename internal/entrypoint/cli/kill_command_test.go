package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestKillCommand_command_PassesRequestToRunner(t *testing.T) {
	// Given
	flags := rootFlags{includeRoot: new(true), allowRoot: new(false)}
	expectedRequest := inbound.KillRequest{
		Patterns: []string{"chrome", "firefox"},
		Process:  inbound.ProcessOptions{IncludeRoot: new(true), AllowRoot: new(false)},
	}
	runner := &MockKillRunner{}
	runner.On("Run", expectedRequest).Return(nil)
	cmd := newKillCommand(runner, &flags).command()

	// When
	err := executeCommand(cmd, []string{"chrome", "firefox"})

	// Then
	require.NoError(t, err)
	runner.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestKillCommand_command_ReturnsErrorOfRunnerUnchanged(t *testing.T) {
	// Given
	failure := errors.New("killing failed")
	runner := &MockKillRunner{}
	runner.On("Run", inbound.KillRequest{Patterns: []string{"proc"}}).Return(failure)
	cmd := newKillCommand(runner, &rootFlags{}).command()

	// When
	err := executeCommand(cmd, []string{"proc"})

	// Then
	assert.Same(t, failure, err)
}
