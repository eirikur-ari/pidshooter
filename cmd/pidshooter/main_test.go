package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/entrypoint/cli"
)

func TestExitCode_MapsErrorToExitStatus(t *testing.T) {
	tests := newExitCodeTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			code := exitCode(test.err)

			// Then
			assert.Equal(t, test.expected, code)
		})
	}
}

func newExitCodeTestCase() []struct {
	name     string
	err      error
	expected int
} {
	argumentError := cli.ArgumentError{Cause: errors.New("flag provided but not defined: -unknown")}

	return []struct {
		name     string
		err      error
		expected int
	}{
		{"no error is success", nil, 0},
		{"malformed arguments", argumentError, 2},
		{"malformed arguments wrapped by another error", fmt.Errorf("running: %w", argumentError), 2},
		{"any other error is a failure", errors.New("boom"), 1},
	}
}
