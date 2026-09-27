package apperror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeInReturnsTrueWhenErrHasMatchingCode(t *testing.T) {
	err := NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil)

	assert.True(t, CodeInvalidConfig.In(err))
}

func TestCodeInReturnsFalseWhenErrHasDifferentCode(t *testing.T) {
	err := NewError(CodeGameFailed, SeverityFatal, "game session failed", nil)

	assert.False(t, CodeInvalidConfig.In(err))
}

func TestCodeInReturnsTrueWhenErrWrapsMatchingError(t *testing.T) {
	appErr := NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil)
	err := fmt.Errorf("wrapped: %w", appErr)

	assert.True(t, CodeInvalidConfig.In(err))
}

func TestCodeInReturnsFalseWhenErrIsNotAnError(t *testing.T) {
	assert.False(t, CodeInvalidConfig.In(errors.New("boom")))
}

func TestCodeInReturnsFalseWhenErrIsNil(t *testing.T) {
	assert.False(t, CodeInvalidConfig.In(nil))
}

func TestErrorUnwrapReturnsErrorCause(t *testing.T) {
	cause := errors.New("disk full")
	err := NewError(CodeStoreSaveFailed, SeverityWarning, "score not saved", cause)

	assert.Equal(t, cause, err.Unwrap())
	assert.Same(t, cause, err.Unwrap())
	assert.ErrorIs(t, err, cause)
}

func TestErrorUnwrapReturnsNilWithNoErrorCause(t *testing.T) {
	err := NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil)

	assert.Nil(t, err.Unwrap())
}

func TestErrorNewError(t *testing.T) {
	tests := newErrorTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewError(test.errCode, test.errSeverity, test.errMessage, test.errCause)
			assert.Equal(t, test.expected, err.Error())
		})
	}
}

func newErrorTestCases() []struct {
	name        string
	errCode     Code
	errSeverity Severity
	errMessage  string
	errCause    error
	expected    string
} {
	tests := []struct {
		name        string
		errCode     Code
		errSeverity Severity
		errMessage  string
		errCause    error
		expected    string
	}{
		{
			name:        "with both message and cause",
			errCode:     CodeStoreLoadFailed,
			errSeverity: SeverityWarning,
			errMessage:  "score not loaded",
			errCause:    errors.New("disk full"),
			expected:    "score not loaded: disk full",
		},
		{
			name:        "with message only",
			errCode:     CodeInvalidConfig,
			errSeverity: SeverityFatal,
			errMessage:  "invalid configuration",
			errCause:    nil,
			expected:    "invalid configuration",
		},
		{
			name:        "with cause only",
			errCode:     CodeProcessNotFound,
			errSeverity: SeverityFatal,
			errMessage:  "",
			errCause:    errors.New("no processes found"),
			expected:    "no processes found",
		},
		{
			name:        "with neither message nor cause",
			errCode:     CodeUnknown,
			errSeverity: SeverityUnknown,
			errMessage:  "",
			errCause:    nil,
			expected:    "",
		},
	}
	return tests
}
