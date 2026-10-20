package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestPlayCommand_command_PassesRequestOfTheFlagsSetToRunner(t *testing.T) {
	tests := newPlayRequestTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			runner := &MockPlayRunner{}
			runner.On("Run", test.expectedRequest).Return(nil)
			cmd := newPlayCommand(runner, &rootCommand{}).command()

			// When
			err := executeCommand(cmd, test.args)

			// Then
			require.NoError(t, err)
			runner.AssertExpectations(t)
			runner.AssertNumberOfCalls(t, "Run", 1)
		})
	}
}

func TestPlayCommand_command_ReturnsErrorOfRunnerUnchanged(t *testing.T) {
	// Given
	failure := errors.New("session failed")
	runner := &MockPlayRunner{}
	runner.On("Run", inbound.PlayRequest{Patterns: []string{"proc"}}).Return(failure)
	cmd := newPlayCommand(runner, &rootCommand{}).command()

	// When
	err := executeCommand(cmd, []string{"proc"})

	// Then
	assert.Same(t, failure, err)
}

func TestPlayCommand_toRequest_ReturnsIndependentCopyOfFlagValues(t *testing.T) {
	// Given
	command := newPlayCommand(&MockPlayRunner{}, &rootCommand{})
	cmd := command.command()
	require.NoError(t, cmd.Flags().Parse([]string{"--confirm", "--speed=3.5", "--time=60"}))

	// When
	request := command.toRequest(cmd.Flags(), nil)

	// Then
	assert.NotSame(t, &command.flags.confirm, request.Game.ConfirmMode, "the request must hold its own copy of the value")
	assert.NotSame(t, &command.flags.speed, request.Game.Speed, "the request must hold its own copy of the value")
	assert.NotSame(t, &command.flags.timeLimit, request.Game.TimeLimit, "the request must hold its own copy of the value")
}

func newPlayRequestTestCase() []struct {
	name            string
	args            []string
	expectedRequest inbound.PlayRequest
} {
	return []struct {
		name            string
		args            []string
		expectedRequest inbound.PlayRequest
	}{
		{"no flags", []string{"proc"}, inbound.PlayRequest{Patterns: []string{"proc"}}},
		{
			"confirm",
			[]string{"proc", "--confirm"},
			inbound.PlayRequest{Patterns: []string{"proc"}, Game: inbound.GameOptions{ConfirmMode: new(true)}},
		},
		{
			"confirm explicitly false",
			[]string{"proc", "--confirm=false"},
			inbound.PlayRequest{Patterns: []string{"proc"}, Game: inbound.GameOptions{ConfirmMode: new(false)}},
		},
		{
			"speed",
			[]string{"proc", "--speed=3.5"},
			inbound.PlayRequest{Patterns: []string{"proc"}, Game: inbound.GameOptions{Speed: new(3.5)}},
		},
		{
			"speed with its value as a separate argument",
			[]string{"proc", "--speed", "2.5"},
			inbound.PlayRequest{Patterns: []string{"proc"}, Game: inbound.GameOptions{Speed: new(2.5)}},
		},
		{
			"time",
			[]string{"proc", "--time=60"},
			inbound.PlayRequest{Patterns: []string{"proc"}, Game: inbound.GameOptions{TimeLimit: new(60)}},
		},
		{
			"all flags",
			[]string{"proc", "--confirm", "--speed=4.0", "--time=90"},
			inbound.PlayRequest{
				Patterns: []string{"proc"},
				Game:     inbound.GameOptions{ConfirmMode: new(true), Speed: new(4.0), TimeLimit: new(90)},
			},
		},
		{
			"flags between patterns",
			[]string{"chrome", "--confirm", "firefox", "--speed=3.5"},
			inbound.PlayRequest{
				Patterns: []string{"chrome", "firefox"},
				Game:     inbound.GameOptions{ConfirmMode: new(true), Speed: new(3.5)},
			},
		},
		{
			"pattern after the end of flags",
			[]string{"--confirm", "--", "--proc"},
			inbound.PlayRequest{Patterns: []string{"--proc"}, Game: inbound.GameOptions{ConfirmMode: new(true)}},
		},
	}
}
