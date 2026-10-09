package game

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
)

func TestKillFailure_toError_ReturnsKillFailedWarningWithCause(t *testing.T) {
	// Given
	failure := killFailure{Name: "stubborn", PID: 200, Err: assert.AnError}

	// When
	err := failure.toError()

	// Then
	assert.Equal(t, apperror.CodeKillFailed, err.Code)
	assert.Equal(t, apperror.SeverityWarning, err.Severity)
	assert.EqualError(t, err, "could not kill stubborn (PID 200): "+assert.AnError.Error())
	assert.ErrorIs(t, err, assert.AnError)
}
