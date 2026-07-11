package process

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_TooShort(t *testing.T) {
	for _, p := range []string{"", "a", "ab"} {
		assert.Error(t, Validate(p), "expected error for pattern %q shorter than MinPatternLength", p)
	}
}

func TestValidate_ExactMinLength(t *testing.T) {
	p := strings.Repeat("a", MinPatternLength)
	assert.NoError(t, Validate(p), "expected no error for min-length pattern")
}

func TestValidate_Valid(t *testing.T) {
	assert.NoError(t, Validate("firefox"))
}

func TestValidate_ExactMaxLength(t *testing.T) {
	p := strings.Repeat("a", MaxPatternLength)
	assert.NoError(t, Validate(p), "expected no error for max-length pattern")
}

func TestValidate_TooLong(t *testing.T) {
	p := strings.Repeat("a", MaxPatternLength+1)
	assert.Error(t, Validate(p), "expected error for pattern exceeding MaxPatternLength")
}

func TestErrNoPatterns_IsSentinel(t *testing.T) {
	require.ErrorIs(t, ErrNoPatterns, ErrNoPatterns, "ErrNoPatterns must satisfy errors.Is against itself")
}
