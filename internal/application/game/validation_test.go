package game

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestValidateSearchPatternsEmptySlice(t *testing.T) {
	err := validateSearchPatterns([]string{})
	require.Error(t, err)
	assert.ErrorIs(t, err, process.ErrNoPatterns)
}

func TestValidateSearchPatternsEmptyTerm(t *testing.T) {
	assert.Error(t, validateSearchPatterns([]string{""}))
}

func TestValidateSearchPatternsTooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		assert.Error(t, validateSearchPatterns([]string{p}), "expected error for pattern %q shorter than MinPatternLength", p)
	}
}

func TestValidateSearchPatternsExactMinLength(t *testing.T) {
	min := strings.Repeat("a", process.MinPatternLength)
	assert.NoError(t, validateSearchPatterns([]string{min}))
}

func TestValidateSearchPatternsTooLong(t *testing.T) {
	long := strings.Repeat("a", process.MaxPatternLength+1)
	assert.Error(t, validateSearchPatterns([]string{long}))
}

func TestValidateSearchPatternsValid(t *testing.T) {
	assert.NoError(t, validateSearchPatterns([]string{"myapp", "worker"}))
}

func TestValidateSearchPatternsExactMaxLength(t *testing.T) {
	exact := strings.Repeat("a", process.MaxPatternLength)
	assert.NoError(t, validateSearchPatterns([]string{exact}))
}
