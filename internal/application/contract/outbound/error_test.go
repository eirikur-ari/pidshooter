package outbound

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorruptedDataErrorWithoutMessage(t *testing.T) {
	assert.Equal(t, "corrupted data", CorruptedDataError{}.Error())
}

func TestCorruptedDataErrorWithMessagePrefixesIt(t *testing.T) {
	err := CorruptedDataError{Message: "score file is too large"}
	assert.Equal(t, "score file is too large: corrupted data", err.Error())
}

func TestCorruptedDataErrorMatchesRegardlessOfMessage(t *testing.T) {
	err := error(CorruptedDataError{Message: "score file is too large"})
	assert.True(t, errors.As(err, &CorruptedDataError{}))
}