package cli

import (
	"bytes"
	"io"
	"strings"
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
			root := newRootCommand(testBuildVersion, &MockPlayRunner{}, &MockListRunner{}, &MockKillRunner{})
			flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
			root.defineFlags(flagSet)
			require.NoError(t, flagSet.Parse(test.args))

			// When
			options := root.toProcessOptions(flagSet)

			// Then
			assert.Equal(t, test.expected, options)
		})
	}
}

func TestRootCommand_cmd_ReturnsArgumentErrorWithoutRunningAnythingWhenArgumentsAreMalformed(t *testing.T) {
	tests := newMalformedArgumentsTestCase()

	for _, test := range tests {
		t.Run(strings.Join(test, " "), func(t *testing.T) {
			// Given
			play, list, kill := &MockPlayRunner{}, &MockListRunner{}, &MockKillRunner{}
			cmd := newRootCommand(testBuildVersion, play, list, kill).cmd

			// When
			err := executeCommand(cmd, test)

			// Then
			require.Error(t, err)
			assert.ErrorAs(t, err, &ArgumentError{})
			assertNoRunnerRan(t, play, list, kill)
		})
	}
}

func TestRootCommand_cmd_SuggestsCommandsTheUnknownCommandResembles(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"typo of a command", []string{"plya", "x"}, "did you mean play?"},
		{"no resemblance", []string{"zzzzzz"}, "unknown command: zzzzzz"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			cmd := newRootCommand(testBuildVersion, &MockPlayRunner{}, &MockListRunner{}, &MockKillRunner{}).cmd

			// When
			err := executeCommand(cmd, test.args)

			// Then
			require.Error(t, err)
			assert.ErrorAs(t, err, &ArgumentError{})
			assert.EqualError(t, err, test.expected)
		})
	}
}

func TestRootCommand_cmd_PrintsHelpWithoutRunningAnything(t *testing.T) {
	tests := newHelpTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			play, list, kill := &MockPlayRunner{}, &MockListRunner{}, &MockKillRunner{}
			cmd := newRootCommand(testBuildVersion, play, list, kill).cmd
			var out bytes.Buffer
			cmd.SetArgs(test.args)
			cmd.SetOut(&out)
			cmd.SetErr(io.Discard)

			// When
			err := cmd.Execute()

			// Then
			require.NoError(t, err)
			assert.Contains(t, out.String(), test.expected)
			assertNoRunnerRan(t, play, list, kill)
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
	cmd := newRootCommand(testBuildVersion, runner, &MockListRunner{}, &MockKillRunner{}).cmd

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
	cmd := newRootCommand(testBuildVersion, &MockPlayRunner{}, runner, &MockKillRunner{}).cmd

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
	cmd := newRootCommand(testBuildVersion, &MockPlayRunner{}, &MockListRunner{}, runner).cmd

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
	cmd := newRootCommand(testBuildVersion, runner, &MockListRunner{}, &MockKillRunner{}).cmd

	// When
	err := executeCommand(cmd, []string{"--include-root", "--i-am-root", "play", "proc"})

	// Then
	require.NoError(t, err)
	runner.AssertExpectations(t)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func newProcessOptionsTestCase() []struct {
	name     string
	args     []string
	expected inbound.ProcessOptions
} {
	return []struct {
		name     string
		args     []string
		expected inbound.ProcessOptions
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

func newMalformedArgumentsTestCase() [][]string {
	return [][]string{
		{"play", "proc", "--unknown"},
		{"list", "proc", "--unknown"},
		{"kill", "proc", "--unknown"},
		{"list", "proc", "--speed=3"}, // a flag of play, not of list
		{"kill", "proc", "--confirm"}, // a flag of play, not of kill
		{"play", "proc", "--speed"},   // missing its value
		{"play", "proc", "--speed=abc"},
		{"play", "proc", "--time=abc"},
		{"play"},
		{"list"},
		{"kill"},
		{"--include-root", "play"},
		{"chrome"},
		{"chrome", "play"},
		{"Play", "chrome"},
		{"help"},
		{"help", "play"},
	}
}

func newHelpTestCase() []struct {
	name     string
	args     []string
	expected string
} {
	return []struct {
		name     string
		args     []string
		expected string
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
