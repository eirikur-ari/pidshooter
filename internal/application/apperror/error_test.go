package apperror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCode_In_ReportsWhetherErrIsAnErrorWithCode(t *testing.T) {
	nested := NewError(CodeKillFailed, SeverityWarning, "could not kill target",
		NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil))

	tests := []struct {
		name     string
		err      error
		code     Code
		expected bool
	}{
		{"matching code", NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil), CodeInvalidConfig, true},
		{"different code", NewError(CodeGameFailed, SeverityFatal, "game session failed", nil), CodeInvalidConfig, false},
		{"wrapped matching error", fmt.Errorf("wrapped: %w", NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil)), CodeInvalidConfig, true},
		{"plain error", errors.New("boom"), CodeInvalidConfig, false},
		{"nil error", nil, CodeInvalidConfig, false},
		{"outer code of nested errors", nested, CodeKillFailed, true},
		{"cause code of nested errors", nested, CodeInvalidConfig, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := test.code.In(test.err)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestNewError_SetsFields(t *testing.T) {
	// Given
	cause := errors.New("disk full")

	// When
	err := NewError(CodeStoreSaveFailed, SeverityWarning, "score not saved", cause)

	// Then
	assert.Equal(t, CodeStoreSaveFailed, err.Code)
	assert.Equal(t, SeverityWarning, err.Severity)
	assert.Equal(t, "score not saved", err.message)
	assert.Same(t, cause, err.cause)
	assert.False(t, err.logged)
}

func TestError_Error_FormatsMessageAndCause(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		cause    error
		expected string
	}{
		{"message and cause", "score not loaded", errors.New("disk full"), "score not loaded: disk full"},
		{"message only", "invalid configuration", nil, "invalid configuration"},
		{"cause only", "", errors.New("no processes found"), "no processes found"},
		{"neither message nor cause", "", nil, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			err := NewError(CodeUnknown, SeverityUnknown, test.message, test.cause)

			// When
			actual := err.Error()

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestError_Unwrap_ReturnsCause(t *testing.T) {
	// Given
	cause := errors.New("disk full")
	err := NewError(CodeStoreSaveFailed, SeverityWarning, "score not saved", cause)

	// When
	unwrapped := err.Unwrap()

	// Then
	assert.Same(t, cause, unwrapped)
	assert.ErrorIs(t, err, cause)
}

func TestError_Unwrap_ReturnsNilWhenThereIsNoCause(t *testing.T) {
	// Given
	err := NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil)

	// When
	unwrapped := err.Unwrap()

	// Then
	assert.Nil(t, unwrapped)
}
