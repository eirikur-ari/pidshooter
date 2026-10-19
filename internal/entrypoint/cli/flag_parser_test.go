package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestFlagParser_parsePlay_ReturnsPlayRequest(t *testing.T) {
	// Given
	parser := newFlagParser()

	// When
	request, err := parser.parsePlay([]string{"chrome", "--confirm", "firefox", "--speed=3.5"})

	// Then
	require.NoError(t, err)
	assert.Equal(t, inbound.PlayRequest{
		Patterns: []string{"chrome", "firefox"},
		Game:     inbound.GameOptions{ConfirmMode: new(true), Speed: new(3.5)},
	}, request)
}

func TestFlagParser_parsePlay_ReturnsArgumentErrorWhenArgumentsAreMalformed(t *testing.T) {
	tests := [][]string{
		{"proc", "--unknown"},
		{"proc", "--speed"},     // missing its value
		{"proc", "--speed=abc"}, // wrong type
		{"proc", "--time=abc"},  // wrong type
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			// Given
			parser := newFlagParser()

			// When
			request, err := parser.parsePlay(args)

			// Then
			require.Error(t, err)
			assert.True(t, errors.As(err, &ArgumentError{}))
			assert.Equal(t, inbound.PlayRequest{}, request)
		})
	}
}

func TestFlagParser_parsePlay_ReturnsArgumentErrorWhenParsingLeavesUnexpectedArgument(t *testing.T) {
	// Given
	parser := newFlagParser()
	parser.splitter = &FakeSplitter{FlagArgs: []string{"leftover"}}

	// When
	request, err := parser.parsePlay([]string{"firefox"})

	// Then
	require.Error(t, err)
	assert.True(t, errors.As(err, &ArgumentError{}))
	assert.EqualError(t, err, "unexpected argument: leftover")
	assert.Equal(t, inbound.PlayRequest{}, request)
}

func TestFlagParser_parseList_ReturnsListRequest(t *testing.T) {
	// Given
	parser := newFlagParser()

	// When
	request, err := parser.parseList([]string{"chrome", "firefox"})

	// Then
	require.NoError(t, err)
	assert.Equal(t, inbound.ListRequest{Patterns: []string{"chrome", "firefox"}}, request)
}

func TestFlagParser_parseList_ReturnsArgumentErrorWhenArgumentsAreFlags(t *testing.T) {
	// Given
	parser := newFlagParser()

	// When
	request, err := parser.parseList([]string{"proc", "--unknown"})

	// Then
	require.Error(t, err)
	assert.True(t, errors.As(err, &ArgumentError{}))
	assert.Equal(t, inbound.ListRequest{}, request)
}

func TestFlagParser_parseKill_ReturnsKillRequest(t *testing.T) {
	// Given
	parser := newFlagParser()

	// When
	request, err := parser.parseKill([]string{"chrome", "firefox"})

	// Then
	require.NoError(t, err)
	assert.Equal(t, inbound.KillRequest{Patterns: []string{"chrome", "firefox"}}, request)
}

func TestFlagParser_parseKill_ReturnsArgumentErrorWhenArgumentsAreFlags(t *testing.T) {
	// Given
	parser := newFlagParser()

	// When
	request, err := parser.parseKill([]string{"proc", "--unknown"})

	// Then
	require.Error(t, err)
	assert.True(t, errors.As(err, &ArgumentError{}))
	assert.Equal(t, inbound.KillRequest{}, request)
}

func TestFlagParser_parseRequest_ReturnsRequestOfCommand(t *testing.T) {
	// Given
	parser := newFlagParser()

	// When
	play, playErr := parser.parseRequest(newCommand([]string{"play", "proc", "--confirm"}))
	list, listErr := parser.parseRequest(newCommand([]string{"list", "proc"}))
	kill, killErr := parser.parseRequest(newCommand([]string{"kill", "proc"}))

	// Then
	require.NoError(t, playErr)
	require.NoError(t, listErr)
	require.NoError(t, killErr)
	assert.Equal(t, inbound.RunRequest{Play: &inbound.PlayRequest{
		Patterns: []string{"proc"},
		Game:     inbound.GameOptions{ConfirmMode: new(true)},
	}}, play)
	assert.Equal(t, inbound.RunRequest{List: &inbound.ListRequest{Patterns: []string{"proc"}}}, list)
	assert.Equal(t, inbound.RunRequest{Kill: &inbound.KillRequest{Patterns: []string{"proc"}}}, kill)
}

func TestFlagParser_parseRequest_ReturnsArgumentErrorWhenNotACommand(t *testing.T) {
	// Given
	parser := newFlagParser()

	// When
	request, err := parser.parseRequest(newCommand([]string{"chrome"}))

	// Then
	require.Error(t, err)
	assert.True(t, errors.As(err, &ArgumentError{}))
	assert.Equal(t, inbound.RunRequest{}, request)
}
