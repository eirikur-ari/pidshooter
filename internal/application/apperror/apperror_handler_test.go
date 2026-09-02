package apperror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func TestHandlerHandleWarningAbsorbsAndLogsAsWarning(t *testing.T) {
	logger := &fake.Logger{}
	h := NewHandler(logger)
	input := NewError(CodeScoreLoadFailed, SeverityWarning, "could not load scores", errors.New("disk full"))

	err := h.Handle(input)

	assert.NoError(t, err)
	assert.Equal(t, []string{"could not load scores: disk full"}, logger.Warnings)
	assert.Empty(t, logger.Errors)
}

func TestHandlerHandleErrorAbsorbsAndLogsAsError(t *testing.T) {
	logger := &fake.Logger{}
	h := NewHandler(logger)
	input := NewError(CodeNoProcessesFound, SeverityError, "no processes found", nil)

	err := h.Handle(input)

	assert.NoError(t, err)
	assert.Equal(t, []string{"no processes found"}, logger.Errors)
	assert.Empty(t, logger.Warnings)
}

func TestHandlerHandleFatalLogsAsErrorAndReturns(t *testing.T) {
	logger := &fake.Logger{}
	h := NewHandler(logger)
	input := NewError(CodeGameFailed, SeverityFatal, "game session failed", errors.New("boom"))

	err := h.Handle(input)

	require.Error(t, err)
	assert.Same(t, input, err)
	assert.Equal(t, []string{"game session failed: boom"}, logger.Errors)
	assert.Empty(t, logger.Warnings)
}

func TestHandlerHandleUnknownSeverityLogsAsErrorAndAbsorbs(t *testing.T) {
	logger := &fake.Logger{}
	h := NewHandler(logger)
	input := errors.New("plain error")

	err := h.Handle(input)

	assert.NoError(t, err)
	assert.Equal(t, []string{"plain error"}, logger.Errors)
	assert.Empty(t, logger.Warnings)
}

func TestHandlerHandleWrappedFatalErrorStillPropagates(t *testing.T) {
	logger := &fake.Logger{}
	h := NewHandler(logger)
	fatal := NewError(CodeGameFailed, SeverityFatal, "game session failed", errors.New("boom"))
	wrapped := fmt.Errorf("during Play: %w", fatal)

	err := h.Handle(wrapped)

	require.Error(t, err)
	assert.Same(t, wrapped, err, "Handle returns whatever it was given, not a re-derived error")
	assert.Equal(t, []string{"during Play: game session failed: boom"}, logger.Errors)
	assert.Empty(t, logger.Warnings)
}

func TestHandlerHandleWrappedWarningStillAbsorbs(t *testing.T) {
	logger := &fake.Logger{}
	h := NewHandler(logger)
	warning := NewError(CodeScoreLoadFailed, SeverityWarning, "score not loaded", errors.New("disk full"))
	wrapped := fmt.Errorf("during LoadScoreBoard: %w", warning)

	err := h.Handle(wrapped)

	assert.NoError(t, err)
	assert.Equal(t, []string{"during LoadScoreBoard: score not loaded: disk full"}, logger.Warnings)
	assert.Empty(t, logger.Errors)
}

func TestHandlerHandleNilErrorReturnsNil(t *testing.T) {
	logger := &fake.Logger{}
	h := NewHandler(logger)

	err := h.Handle(nil)

	assert.NoError(t, err)
	assert.Empty(t, logger.Errors)
	assert.Empty(t, logger.Warnings)
}