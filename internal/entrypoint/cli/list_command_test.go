package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestListCommand_command_PassesRequestToRunner(t *testing.T) {
	// Given
	runner := &MockListRunner{}
	runner.On("Run", inbound.ListRequest{Patterns: []string{"chrome", "firefox"}}).Return(nil)
	cmd := newListCommand(runner, &rootCommand{}).command()

	// When
	err := executeCommand(cmd, []string{"chrome", "firefox"})

	// Then
	require.NoError(t, err)
	runner.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestListCommand_command_ReturnsErrorOfRunnerUnchanged(t *testing.T) {
	// Given
	failure := errors.New("listing failed")
	runner := &MockListRunner{}
	runner.On("Run", inbound.ListRequest{Patterns: []string{"proc"}}).Return(failure)
	cmd := newListCommand(runner, &rootCommand{}).command()

	// When
	err := executeCommand(cmd, []string{"proc"})

	// Then
	assert.Same(t, failure, err)
}
