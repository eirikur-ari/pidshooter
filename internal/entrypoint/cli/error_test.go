package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArgumentErrorReturnsWrappedMessage(t *testing.T) {
	err := ArgumentError{Cause: errors.New("flag provided but not defined: -unknown")}
	assert.Equal(t, "flag provided but not defined: -unknown", err.Error())
}

func TestArgumentErrorMatchesRegardlessOfCause(t *testing.T) {
	err := error(ArgumentError{Cause: errors.New("boom")})
	assert.True(t, errors.As(err, &ArgumentError{}))
}

func TestArgumentErrorUnwrapsToCause(t *testing.T) {
	cause := errors.New("boom")
	err := ArgumentError{Cause: cause}
	assert.Same(t, cause, err.Unwrap())
}