package logger

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger_WritesToStandardError(t *testing.T) {
	// When
	logger := NewLogger()

	// Then
	assert.Same(t, os.Stderr, logger.writer)
}

func TestLogger_Warn_WritesWarningLine(t *testing.T) {
	tests := newWarnTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			var output bytes.Buffer
			logger := &Logger{writer: &output}

			// When
			logger.Warn(test.message)
			actual := output.String()

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestLogger_Error_WritesErrorLine(t *testing.T) {
	tests := newErrorTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			var output bytes.Buffer
			logger := &Logger{writer: &output}

			// When
			logger.Error(test.message)
			actual := output.String()

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func newWarnTestCases() []struct {
	name     string
	message  string
	expected string
} {
	return []struct {
		name     string
		message  string
		expected string
	}{
		{"message", "config file ignored", "warning: config file ignored\n"},
		{"empty message", "", "warning: \n"},
	}
}

func newErrorTestCases() []struct {
	name     string
	message  string
	expected string
} {
	return []struct {
		name     string
		message  string
		expected string
	}{
		{"message", "could not save score", "error: could not save score\n"},
		{"empty message", "", "error: \n"},
	}
}
