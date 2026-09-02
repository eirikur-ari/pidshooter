package apperror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorErrorWithMessageAndCauseJoinsBoth(t *testing.T) {
	err := NewError(CodeScoreLoadFailed, SeverityWarning, "score not loaded", errors.New("disk full"))

	assert.Equal(t, "score not loaded: disk full", err.Error())
}

func TestErrorErrorWithEmptyMessageReturnsCauseAlone(t *testing.T) {
	cause := errors.New("no processes found")
	err := NewError(CodeNoProcessesFound, SeverityFatal, "", cause)

	assert.Equal(t, "no processes found", err.Error())
}

func TestErrorErrorWithNoCauseReturnsMessageAlone(t *testing.T) {
	err := NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil)

	assert.Equal(t, "invalid configuration", err.Error())
}

func TestErrorUnwrapReturnsCause(t *testing.T) {
	cause := errors.New("disk full")
	err := NewError(CodeScoreSaveFailed, SeverityWarning, "score not saved", cause)

	assert.Same(t, cause, err.Unwrap())
}

func TestErrorUnwrapReturnsNilWithNoCause(t *testing.T) {
	err := NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil)

	assert.Nil(t, err.Unwrap())
}

func TestErrorIsTraversesToCause(t *testing.T) {
	cause := errors.New("disk full")
	err := NewError(CodeScoreSaveFailed, SeverityWarning, "score not saved", cause)

	assert.ErrorIs(t, err, cause)
}