package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestFlagParser_parse_ReturnsPatternsAndOptions(t *testing.T) {
	// Given
	parser := newFlagParser()

	// When
	patterns, options, err := parser.parse([]string{"chrome", "--confirm", "firefox", "--speed=3.5"})

	// Then
	require.NoError(t, err)
	assert.Equal(t, []string{"chrome", "firefox"}, patterns)
	assert.Equal(t, config.Options{
		Game: config.GameOptions{ConfirmMode: testutil.Pointer(true), Speed: testutil.Pointer(3.5)},
	}, options)
}

func TestFlagParser_parse_ReturnsArgumentErrorWhenArgumentsAreMalformed(t *testing.T) {
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
			patterns, options, err := parser.parse(args)

			// Then
			require.Error(t, err)
			assert.True(t, errors.As(err, &ArgumentError{}))
			assert.Nil(t, patterns)
			assert.Equal(t, config.Options{}, options)
		})
	}
}

func TestFlagParser_parse_ReturnsArgumentErrorWhenParsingLeavesUnexpectedArgument(t *testing.T) {
	// Given
	parser := newFlagParser()
	parser.splitter = &FakeSplitter{FlagArgs: []string{"leftover"}}

	// When
	patterns, options, err := parser.parse([]string{"firefox"})

	// Then
	require.Error(t, err)
	assert.True(t, errors.As(err, &ArgumentError{}))
	assert.EqualError(t, err, "unexpected argument: leftover")
	assert.Nil(t, patterns)
	assert.Equal(t, config.Options{}, options)
}
