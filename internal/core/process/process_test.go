package process

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePatternsNoPatterns(t *testing.T) {
	assert.ErrorContains(t, ValidatePatterns(nil), "at least one search pattern is required")
	assert.ErrorContains(t, ValidatePatterns([]string{}), "at least one search pattern is required")
}

func TestValidatePatternsTooShort(t *testing.T) {
	for _, p := range []string{"", "a", "ab"} {
		assert.Error(t, ValidatePatterns([]string{p}), "expected error for pattern %q shorter than MinPatternLength", p)
	}
}

func TestValidatePatternsExactMinLength(t *testing.T) {
	p := strings.Repeat("a", MinPatternLength)
	assert.NoError(t, ValidatePatterns([]string{p}), "expected no error for min-length pattern")
}

func TestValidatePatternsValid(t *testing.T) {
	assert.NoError(t, ValidatePatterns([]string{"firefox"}))
}

func TestValidatePatternsExactMaxLength(t *testing.T) {
	p := strings.Repeat("a", MaxPatternLength)
	assert.NoError(t, ValidatePatterns([]string{p}), "expected no error for max-length pattern")
}

func TestValidatePatternsTooLong(t *testing.T) {
	p := strings.Repeat("a", MaxPatternLength+1)
	assert.Error(t, ValidatePatterns([]string{p}), "expected error for pattern exceeding MaxPatternLength")
}

func TestValidateProcessesEmpty(t *testing.T) {
	assert.ErrorContains(t, ValidateProcesses(nil), "no processes found")
}

func TestValidateProcessesNonEmpty(t *testing.T) {
	assert.NoError(t, ValidateProcesses([]Info{NewInfo(100, "target", 4096)}))
}

func TestValidateNameMismatch(t *testing.T) {
	err := ValidateName("target", "somethingElse")
	assert.ErrorContains(t, err, `expected "target"`)
	assert.ErrorContains(t, err, `got "somethingElse"`)
}

func TestValidateNameMatch(t *testing.T) {
	assert.NoError(t, ValidateName("target", "target"))
}

func TestInfoIsProtected(t *testing.T) {
	assert.True(t, NewInfo(0, "swapper", 0).IsProtected())
	assert.True(t, NewInfo(1, "init", 0).IsProtected())
	assert.False(t, NewInfo(2, "init", 0).IsProtected())
	assert.False(t, NewInfo(100, "myapp", 0).IsProtected())
}

func TestFindMatchesByName(t *testing.T) {
	infos := []Info{
		NewInfo(100, "myapp", 1024),
		NewInfo(200, "worker", 2048),
	}
	result := Find(infos, []string{"myapp"}, 0)
	require.Len(t, result, 1)
	assert.Equal(t, 100, result[0].Pid)
}

func TestFindSubstringMatch(t *testing.T) {
	infos := []Info{NewInfo(100, "myapp-worker", 1024)}
	result := Find(infos, []string{"app"}, 0)
	assert.Len(t, result, 1)
}

func TestFindCaseInsensitive(t *testing.T) {
	infos := []Info{NewInfo(100, "MyApp", 1024)}
	result := Find(infos, []string{"myapp"}, 0)
	assert.Len(t, result, 1)
}

func TestFindMultipleTerms(t *testing.T) {
	infos := []Info{
		NewInfo(100, "myapp", 1024),
		NewInfo(200, "worker", 2048),
		NewInfo(300, "other", 512),
	}
	result := Find(infos, []string{"myapp", "worker"}, 0)
	assert.Len(t, result, 2)
}

func TestFindNoMatch(t *testing.T) {
	infos := []Info{NewInfo(100, "myapp", 1024)}
	result := Find(infos, []string{"worker"}, 0)
	assert.Empty(t, result)
}

func TestFindExcludesPID1(t *testing.T) {
	infos := []Info{
		NewInfo(1, "init", 512),
		NewInfo(100, "myapp", 1024),
	}
	result := Find(infos, []string{"init", "myapp"}, 0)
	for _, p := range result {
		assert.NotEqual(t, 1, p.Pid, "should not include PID 1")
	}
}

func TestFindExcludesOwnPID(t *testing.T) {
	const ownPID = 999
	infos := []Info{
		NewInfo(ownPID, "testprocess", 1024),
		NewInfo(100, "testprocess", 2048),
	}
	result := Find(infos, []string{"testprocess"}, ownPID)
	for _, p := range result {
		assert.NotEqual(t, ownPID, p.Pid, "should not include own PID")
	}
	require.Len(t, result, 1)
	assert.Equal(t, 100, result[0].Pid)
}

func TestFindExcludesPID0(t *testing.T) {
	infos := []Info{
		NewInfo(0, "swapper", 0),
		NewInfo(100, "myapp", 1024),
	}
	result := Find(infos, []string{"swapper", "myapp"}, -1)
	for _, p := range result {
		assert.Greater(t, p.Pid, 1, "filter should exclude PID %d", p.Pid)
	}
	require.Len(t, result, 1)
	assert.Equal(t, 100, result[0].Pid)
}
