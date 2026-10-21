package apperror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError_SetsFields(t *testing.T) {
	// Given
	cause := errors.New("disk full")

	// When
	err := NewError(CodeStoreFailed, SeverityWarning, "score not saved", cause)

	// Then
	assert.Equal(t, CodeStoreFailed, err.Code)
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
	err := NewError(CodeStoreFailed, SeverityWarning, "score not saved", cause)

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

func TestError_severity_ReturnsSeverity(t *testing.T) {
	// Given
	err := NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil)

	// When
	actual := err.severity()

	// Then
	assert.Equal(t, SeverityFatal, actual)
}

func TestError_severity_ReturnsUnknownForNilError(t *testing.T) {
	// Given
	var err *Error

	// When
	actual := err.severity()

	// Then
	assert.Equal(t, SeverityUnknown, actual)
}
