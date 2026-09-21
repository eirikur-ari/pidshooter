package apperror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorUnwrapReturnsErrorCause(t *testing.T) {
	cause := errors.New("disk full")
	err := NewError(CodeScoreSaveFailed, SeverityWarning, "score not saved", cause)

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
			errCode:     CodeScoreLoadFailed,
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

func TestHandleAbsorbsWarningSeverity(t *testing.T) {
	input := NewError(CodeScoreLoadFailed, SeverityWarning, "could not load scores", errors.New("disk full"))

	err := Handle(input)

	assert.NoError(t, err)
}

func TestHandleAbsorbsErrorSeverity(t *testing.T) {
	input := NewError(CodeProcessNotFound, SeverityError, "no processes found", nil)

	err := Handle(input)

	assert.NoError(t, err)
}

func TestHandleReturnsFatalSeverityError(t *testing.T) {
	input := NewError(CodeGameFailed, SeverityFatal, "game session failed", errors.New("boom"))

	err := Handle(input)

	require.Error(t, err)
	assert.Same(t, input, err)
}

func TestHandlePropagatesUnknownSeverity(t *testing.T) {
	input := errors.New("plain error")

	err := Handle(input)

	require.Error(t, err)
	assert.Same(t, input, err)
}

func TestHandleLogsFatalErrorOnlyOnce(t *testing.T) {
	fatal := NewError(CodeGameFailed, SeverityFatal, "game session failed", errors.New("boom"))
	assert.False(t, fatal.logged, "logged should start false")

	first := Handle(fatal)
	assert.True(t, fatal.logged, "logged should be set after the first Handle call")

	second := Handle(fatal)

	require.Error(t, first)
	require.Error(t, second)
	assert.Same(t, fatal, first)
	assert.Same(t, fatal, second)
}

func TestHandleReturnsWrappedFatalError(t *testing.T) {
	fatal := NewError(CodeGameFailed, SeverityFatal, "game session failed", errors.New("boom"))
	wrapped := fmt.Errorf("during Play: %w", fatal)

	err := Handle(wrapped)

	require.Error(t, err)
	assert.Same(t, wrapped, err, "Handle returns whatever it was given, not a re-derived error")
}

func TestHandleAbsorbsWrappedWarning(t *testing.T) {
	warning := NewError(CodeScoreLoadFailed, SeverityWarning, "score not loaded", errors.New("disk full"))
	wrapped := fmt.Errorf("during LoadScoreBoard: %w", warning)

	err := Handle(wrapped)

	assert.NoError(t, err)
}

func TestHandleReturnsNilErrorWhenGivenNil(t *testing.T) {
	err := Handle(nil)

	assert.NoError(t, err)
}
