package apperror

import (
	"errors"
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