package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/entrypoint/cli"
)

func TestExitCodeReturnsZeroForNil(t *testing.T) {
	assert.Equal(t, 0, exitCode(nil))
}

func TestExitCodeReturnsTwoForArgumentError(t *testing.T) {
	err := cli.ArgumentError{Cause: errors.New("flag provided but not defined: -unknown")}
	assert.Equal(t, 2, exitCode(err))
}

func TestExitCodeReturnsTwoForWrappedInvalidConfig(t *testing.T) {
	invalid := apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", errors.New("boom"))
	err := cli.ArgumentError{Cause: invalid}
	assert.Equal(t, 2, exitCode(err))
}

func TestExitCodeReturnsOneForUnwrappedInvalidConfig(t *testing.T) {
	err := apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", errors.New("boom"))
	assert.Equal(t, 1, exitCode(err))
}

func TestExitCodeReturnsOneForOtherErrors(t *testing.T) {
	err := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	assert.Equal(t, 1, exitCode(err))
}

func TestExitCodeReturnsOneForPlainError(t *testing.T) {
	assert.Equal(t, 1, exitCode(errors.New("boom")))
}
