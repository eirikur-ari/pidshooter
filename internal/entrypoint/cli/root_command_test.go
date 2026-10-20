package cli

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestRootCommand_toProcessOptions_ReturnsOnlyTheFlagsSet(t *testing.T) {
	tests := newProcessOptionsTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			root := newRootCommand(&MockPlayRunner{}, &MockListRunner{}, &MockKillRunner{})
			flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
			root.defineFlags(flagSet)
			require.NoError(t, flagSet.Parse(test.args))

			// When
			options := root.toProcessOptions(flagSet)

			// Then
			assert.Equal(t, test.expectedOptions, options)
		})
	}
}

func TestRootCommand_command_PassesProcessOptionsToPlayRunner(t *testing.T) {
	// Given
	playRequest := inbound.PlayRequest{
		Patterns: []string{"proc"},
		Process:  inbound.ProcessOptions{IncludeRoot: new(true), AllowRoot: new(false)},
	}
	runner := &MockPlayRunner{}
	runner.On("Run", playRequest).Return(nil)
	cmd := newRootCommand(runner, &MockListRunner{}, &MockKillRunner{}).cmd

	// When
	err := executeCommand(cmd, []string{"play", "proc", "--include-root", "--i-am-root=false"})

	// Then
	require.NoError(t, err)
	runner.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestRootCommand_command_PassesProcessOptionsToListRunner(t *testing.T) {
	// Given
	listRequest := inbound.ListRequest{
		Patterns: []string{"proc"},
		Process:  inbound.ProcessOptions{IncludeRoot: new(true), AllowRoot: new(false)},
	}
	runner := &MockListRunner{}
	runner.On("Run", listRequest).Return(nil)
	cmd := newRootCommand(&MockPlayRunner{}, runner, &MockKillRunner{}).cmd

	// When
	err := executeCommand(cmd, []string{"list", "proc", "--include-root", "--i-am-root=false"})

	// Then
	require.NoError(t, err)
	runner.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestRootCommand_command_PassesProcessOptionsToKillRunner(t *testing.T) {
	// Given
	killRequest := inbound.KillRequest{
		Patterns: []string{"proc"},
		Process:  inbound.ProcessOptions{IncludeRoot: new(true), AllowRoot: new(false)},
	}
	runner := &MockKillRunner{}
	runner.On("Run", killRequest).Return(nil)
	cmd := newRootCommand(&MockPlayRunner{}, &MockListRunner{}, runner).cmd

	// When
	err := executeCommand(cmd, []string{"kill", "proc", "--include-root", "--i-am-root=false"})

	// Then
	require.NoError(t, err)
	runner.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestRootCommand_command_PassesRootFlagsPlacedBeforeTheCommandToRunner(t *testing.T) {
	// Given
	playRequest := inbound.PlayRequest{
		Patterns: []string{"proc"},
		Process:  inbound.ProcessOptions{IncludeRoot: new(true), AllowRoot: new(true)},
	}
	runner := &MockPlayRunner{}
	runner.On("Run", playRequest).Return(nil)
	cmd := newRootCommand(runner, &MockListRunner{}, &MockKillRunner{}).cmd

	// When
	err := executeCommand(cmd, []string{"--include-root", "--i-am-root", "play", "proc"})

	// Then
	require.NoError(t, err)
	runner.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func newProcessOptionsTestCase() []struct {
	name            string
	args            []string
	expectedOptions inbound.ProcessOptions
} {
	return []struct {
		name            string
		args            []string
		expectedOptions inbound.ProcessOptions
	}{
		{"no flags", nil, inbound.ProcessOptions{}},
		{"include-root", []string{"--include-root"}, inbound.ProcessOptions{IncludeRoot: new(true)}},
		{"include-root explicitly false", []string{"--include-root=false"}, inbound.ProcessOptions{IncludeRoot: new(false)}},
		{"i-am-root", []string{"--i-am-root"}, inbound.ProcessOptions{AllowRoot: new(true)}},
		{
			"both flags",
			[]string{"--include-root", "--i-am-root"},
			inbound.ProcessOptions{IncludeRoot: new(true), AllowRoot: new(true)},
		},
	}
}
