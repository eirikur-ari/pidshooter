package apperror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleAbsorbsWarningSeverity(t *testing.T) {
	input := NewError(CodeScoreLoadFailed, SeverityWarning, "could not load scores", errors.New("disk full"))

	err := newTestHandler().Handle(input)

	assert.NoError(t, err)
}

func TestHandleAbsorbsErrorSeverity(t *testing.T) {
	input := NewError(CodeProcessNotFound, SeverityError, "no processes found", nil)

	err := newTestHandler().Handle(input)

	assert.NoError(t, err)
}

func TestHandleReturnsFatalSeverityError(t *testing.T) {
	input := NewError(CodeGameFailed, SeverityFatal, "game session failed", errors.New("boom"))

	err := newTestHandler().Handle(input)

	require.Error(t, err)
	assert.Same(t, input, err)
}

func TestHandlePropagatesUnknownSeverity(t *testing.T) {
	input := errors.New("plain error")

	err := newTestHandler().Handle(input)

	require.Error(t, err)
	assert.Same(t, input, err)
}

func TestHandleLogsFatalErrorOnlyOnce(t *testing.T) {
	fatal := NewError(CodeGameFailed, SeverityFatal, "game session failed", errors.New("boom"))
	assert.False(t, fatal.logged, "logged should start false")
	handler := newTestHandler()

	first := handler.Handle(fatal)
	assert.True(t, fatal.logged, "logged should be set after the first Handle call")

	second := handler.Handle(fatal)

	require.Error(t, first)
	require.Error(t, second)
	assert.Same(t, fatal, first)
	assert.Same(t, fatal, second)
}

func TestHandleReturnsWrappedFatalError(t *testing.T) {
	fatal := NewError(CodeGameFailed, SeverityFatal, "game session failed", errors.New("boom"))
	wrapped := fmt.Errorf("during Play: %w", fatal)

	err := newTestHandler().Handle(wrapped)

	require.Error(t, err)
	assert.Same(t, wrapped, err, "Handle returns whatever it was given, not a re-derived error")
}

func TestHandleAbsorbsWrappedWarning(t *testing.T) {
	warning := NewError(CodeScoreLoadFailed, SeverityWarning, "score not loaded", errors.New("disk full"))
	wrapped := fmt.Errorf("during LoadScoreBoard: %w", warning)

	err := newTestHandler().Handle(wrapped)

	assert.NoError(t, err)
}

func TestHandleReturnsNilErrorWhenGivenNil(t *testing.T) {
	err := newTestHandler().Handle(nil)

	assert.NoError(t, err)
}

func newTestHandler() *Handler {
	return NewHandler(noopLogger{})
}

// noopLogger discards every message. error_handler_test.go cannot use
// testutil/fake.Logger here: fake also imports application/config, which
// imports apperror, and pulling fake into apperror's white-box test would
// create an import cycle.
type noopLogger struct{}

func (noopLogger) Warn(string)  {}
func (noopLogger) Error(string) {}
