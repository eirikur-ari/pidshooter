package game

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
)

func TestKillDud_toError_ReturnsProcessNotFoundWarning(t *testing.T) {
	// Given
	dud := killDud{Name: "gone", PID: 300}

	// When
	err := dud.toError()

	// Then
	assert.Equal(t, apperror.CodeProcessNotFound, err.Code)
	assert.Equal(t, apperror.SeverityWarning, err.Severity)
	assert.EqualError(t, err, "gone (PID 300) ran away before it could be killed")
}
