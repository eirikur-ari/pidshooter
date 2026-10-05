package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArgumentError_Error_ReturnsCauseMessage(t *testing.T) {
	// Given
	err := ArgumentError{Cause: errors.New("flag provided but not defined: -unknown")}

	// Then
	assert.Equal(t, "flag provided but not defined: -unknown", err.Error())
}

func TestArgumentError_Error_ReturnsGenericMessageWhenCauseIsNil(t *testing.T) {
	// Given
	err := ArgumentError{}

	// Then
	assert.Equal(t, "invalid arguments", err.Error())
}

func TestArgumentError_Unwrap_ReturnsCause(t *testing.T) {
	// Given
	cause := errors.New("boom")

	// When
	err := ArgumentError{Cause: cause}.Unwrap()

	// Then
	assert.Same(t, cause, err)
}

func TestArgumentError_ErrorsAs_MatchesByTypeWhenWrapped(t *testing.T) {
	// Given
	err := fmt.Errorf("context: %w", ArgumentError{Cause: errors.New("boom")})

	// Then
	assert.True(t, errors.As(err, &ArgumentError{}))
}

func TestArgumentError_ErrorsAs_DoesNotMatchOtherErrors(t *testing.T) {
	// Given
	err := errors.New("boom")

	// Then
	assert.False(t, errors.As(err, &ArgumentError{}))
}
