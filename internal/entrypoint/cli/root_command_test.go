package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestRootCommand_mapFlags_StoresOnlyTheFlagsSet(t *testing.T) {
	tests := newRootFlagsTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			root := newRootCommand(&MockPlayRunner{}, &MockListRunner{}, &MockKillRunner{})
			cmd := &cobra.Command{}
			root.defineFlags(cmd.Flags())
			require.NoError(t, cmd.Flags().Parse(test.args))

			// When
			err := root.mapFlags(cmd, nil)

			// Then
			require.NoError(t, err)
			assert.Equal(t, test.expectedFlags, root.flags)
		})
	}
}

func TestRootCommand_mapFlags_ReturnsErrorWhenAFlagCannotBeRead(t *testing.T) {
	tests := []string{"include-root", "i-am-root"}

	for _, flagName := range tests {
		t.Run(flagName, func(t *testing.T) {
			// Given
			root := newRootCommand(&MockPlayRunner{}, &MockListRunner{}, &MockKillRunner{})
			cmd := &cobra.Command{}
			cmd.Flags().String(flagName, "", "")
			require.NoError(t, cmd.Flags().Parse([]string{"--" + flagName + "=value"}))

			// When
			err := root.mapFlags(cmd, nil)

			// Then
			assert.Error(t, err)
			assert.Equal(t, rootFlags{}, root.flags)
		})
	}
}

func TestRootCommand_command_PassesProcessOptionsToPlayRunner(t *testing.T) {
	// Given
	expectedRequest := inbound.PlayRequest{
		Patterns: []string{"proc"},
		Process:  inbound.ProcessOptions{IncludeRoot: new(true), AllowRoot: new(false)},
	}
	runner := &MockPlayRunner{}
	runner.On("Run", expectedRequest).Return(nil)
	cmd := newRootCommand(runner, &MockListRunner{}, &MockKillRunner{}).command()

	// When
	err := executeCommand(cmd, []string{"play", "proc", "--include-root", "--i-am-root=false"})

	// Then
	require.NoError(t, err)
	runner.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestRootCommand_command_PassesProcessOptionsToListRunner(t *testing.T) {
	// Given
	expectedRequest := inbound.ListRequest{
		Patterns: []string{"proc"},
		Process:  inbound.ProcessOptions{IncludeRoot: new(true), AllowRoot: new(false)},
	}
	runner := &MockListRunner{}
	runner.On("Run", expectedRequest).Return(nil)
	cmd := newRootCommand(&MockPlayRunner{}, runner, &MockKillRunner{}).command()

	// When
	err := executeCommand(cmd, []string{"list", "proc", "--include-root", "--i-am-root=false"})

	// Then
	require.NoError(t, err)
	runner.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestRootCommand_command_PassesProcessOptionsToKillRunner(t *testing.T) {
	// Given
	expectedRequest := inbound.KillRequest{
		Patterns: []string{"proc"},
		Process:  inbound.ProcessOptions{IncludeRoot: new(true), AllowRoot: new(false)},
	}
	runner := &MockKillRunner{}
	runner.On("Run", expectedRequest).Return(nil)
	cmd := newRootCommand(&MockPlayRunner{}, &MockListRunner{}, runner).command()

	// When
	err := executeCommand(cmd, []string{"kill", "proc", "--include-root", "--i-am-root=false"})

	// Then
	require.NoError(t, err)
	runner.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func newRootFlagsTestCase() []struct {
	name          string
	args          []string
	expectedFlags rootFlags
} {
	return []struct {
		name          string
		args          []string
		expectedFlags rootFlags
	}{
		{"no flags", nil, rootFlags{}},
		{"include-root", []string{"--include-root"}, rootFlags{includeRoot: new(true)}},
		{"include-root explicitly false", []string{"--include-root=false"}, rootFlags{includeRoot: new(false)}},
		{"i-am-root", []string{"--i-am-root"}, rootFlags{allowRoot: new(true)}},
		{"both flags", []string{"--include-root", "--i-am-root"}, rootFlags{includeRoot: new(true), allowRoot: new(true)}},
	}
}
