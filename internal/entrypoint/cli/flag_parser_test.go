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
