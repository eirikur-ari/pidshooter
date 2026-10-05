package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlagSplitter_split_PartitionsPatternsAndFlagArguments(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedPatterns []string
		expectedFlagArgs []string
	}{
		{"empty args", nil, nil, nil},
		{"all patterns", []string{"chrome", "firefox"}, []string{"chrome", "firefox"}, nil},
		{"pattern then bool flag", []string{"chrome", "--confirm"}, []string{"chrome"}, []string{"--confirm"}},
		{"bool flag then pattern", []string{"--confirm", "chrome"}, []string{"chrome"}, []string{"--confirm"}},
		{"bool flag with single dash", []string{"-confirm", "chrome"}, []string{"chrome"}, []string{"-confirm"}},
		{"bool flag with inline value", []string{"--confirm=false", "chrome"}, []string{"chrome"}, []string{"--confirm=false"}},
		{"bool flag does not consume the next argument", []string{"--confirm", "false"}, []string{"false"}, []string{"--confirm"}},
		{"value flag with inline value", []string{"--speed=3.5", "chrome"}, []string{"chrome"}, []string{"--speed=3.5"}},
		{"value flag with separate value", []string{"--speed", "3.5", "chrome"}, []string{"chrome"}, []string{"--speed", "3.5"}},
		{"value flag with single dash", []string{"-speed", "3.5", "chrome"}, []string{"chrome"}, []string{"-speed", "3.5"}},
		{
			"patterns and flags fully interspersed",
			[]string{"chrome", "--speed", "3.5", "firefox", "--confirm", "node"},
			[]string{"chrome", "firefox", "node"},
			[]string{"--speed", "3.5", "--confirm"},
		},
		{"lone dash is a pattern", []string{"-"}, []string{"-"}, nil},
		{
			"double dash makes the remaining arguments patterns",
			[]string{"chrome", "--", "--confirm", "-suspicious"},
			[]string{"chrome", "--confirm", "-suspicious"},
			nil,
		},
		{"double dash as last argument", []string{"chrome", "--"}, []string{"chrome"}, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			splitter := newFlagSplitter(newFlagSetFixture())

			// When
			patterns, flagArgs, err := splitter.split(test.args)

			// Then
			require.NoError(t, err)
			assert.Equal(t, test.expectedPatterns, patterns)
			assert.Equal(t, test.expectedFlagArgs, flagArgs)
		})
	}
}

func TestFlagSplitter_split_ReturnsErrorWhenArgumentsAreInvalid(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{"unregistered flag", []string{"--bogus"}, "flag provided but not defined: -bogus"},
		{"unregistered flag with single dash", []string{"-bogus"}, "flag provided but not defined: -bogus"},
		{"value flag missing its value", []string{"--speed"}, "flag needs an argument: -speed"},
		{"value flag followed by a flag", []string{"--speed", "--confirm"}, "flag needs an argument: -speed"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			splitter := newFlagSplitter(newFlagSetFixture())

			// When
			patterns, flagArgs, err := splitter.split(test.args)

			// Then
			assert.EqualError(t, err, test.expectedError)
			assert.Nil(t, patterns)
			assert.Nil(t, flagArgs)
		})
	}
}
