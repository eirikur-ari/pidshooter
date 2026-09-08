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
	assert.NoError(t, ValidateProcesses([]Info{NewInfo(100, "target", 4096, 0)}))
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
	assert.True(t, NewInfo(0, "swapper", 0, 0).IsProtected())
	assert.True(t, NewInfo(1, "init", 0, 0).IsProtected())
	assert.False(t, NewInfo(2, "init", 0, 0).IsProtected())
	assert.False(t, NewInfo(100, "myapp", 0, 0).IsProtected())
}

func TestInfoIsKillableBy(t *testing.T) {
	owned := NewInfo(100, "myapp", 0, 1000)
	assert.True(t, owned.IsKillableBy(1000), "caller should be able to kill a process it owns")
	assert.False(t, owned.IsKillableBy(2000), "caller should not be able to kill a process owned by someone else")
	assert.True(t, owned.IsKillableBy(0), "root should be able to kill any process")

	rootOwned := NewInfo(100, "sshd", 0, 0)
	assert.True(t, rootOwned.IsKillableBy(0), "root should be able to kill its own processes")
	assert.False(t, rootOwned.IsKillableBy(1000), "non-root caller should not be able to kill a root-owned process")
}

func TestFindExcludesProcessesNotOwnedByCaller(t *testing.T) {
	infos := []Info{
		NewInfo(100, "myapp", 1024, 1000),
		NewInfo(200, "otherapp", 1024, 2000),
	}
	result := Find(infos, []string{"app"}, 0, 1000)
	require.Len(t, result, 1)
	assert.Equal(t, 100, result[0].PID)
}

func TestFindAsRootIncludesProcessesOwnedByAnyUser(t *testing.T) {
	infos := []Info{
		NewInfo(100, "myapp", 1024, 1000),
		NewInfo(200, "otherapp", 1024, 2000),
	}
	result := Find(infos, []string{"app"}, 0, 0)
	assert.Len(t, result, 2, "root should see processes regardless of owner")
}

func TestFindMatchesByName(t *testing.T) {
	infos := []Info{
		NewInfo(100, "myapp", 1024, 0),
		NewInfo(200, "worker", 2048, 0),
	}
	result := Find(infos, []string{"myapp"}, 0, 0)
	require.Len(t, result, 1)
	assert.Equal(t, 100, result[0].PID)
}

func TestFindSubstringMatch(t *testing.T) {
	infos := []Info{NewInfo(100, "myapp-worker", 1024, 0)}
	result := Find(infos, []string{"app"}, 0, 0)
	assert.Len(t, result, 1)
}

func TestFindCaseInsensitive(t *testing.T) {
	infos := []Info{NewInfo(100, "MyApp", 1024, 0)}
	result := Find(infos, []string{"myapp"}, 0, 0)
	assert.Len(t, result, 1)
}

func TestFindMultipleTerms(t *testing.T) {
	infos := []Info{
		NewInfo(100, "myapp", 1024, 0),
		NewInfo(200, "worker", 2048, 0),
		NewInfo(300, "other", 512, 0),
	}
	result := Find(infos, []string{"myapp", "worker"}, 0, 0)
	assert.Len(t, result, 2)
}

func TestFindNoMatch(t *testing.T) {
	infos := []Info{NewInfo(100, "myapp", 1024, 0)}
	result := Find(infos, []string{"worker"}, 0, 0)
	assert.Empty(t, result)
}

func TestFindExcludesPID1(t *testing.T) {
	infos := []Info{
		NewInfo(1, "init", 512, 0),
		NewInfo(100, "myapp", 1024, 0),
	}
	result := Find(infos, []string{"init", "myapp"}, 0, 0)
	for _, p := range result {
		assert.NotEqual(t, 1, p.PID, "should not include PID 1")
	}
}

func TestFindExcludesOwnPID(t *testing.T) {
	const ownPID = 999
	infos := []Info{
		NewInfo(ownPID, "testprocess", 1024, 0),
		NewInfo(100, "testprocess", 2048, 0),
	}
	result := Find(infos, []string{"testprocess"}, ownPID, 0)
	for _, p := range result {
		assert.NotEqual(t, ownPID, p.PID, "should not include own PID")
	}
	require.Len(t, result, 1)
	assert.Equal(t, 100, result[0].PID)
}

func TestFindExcludesPID0(t *testing.T) {
	infos := []Info{
		NewInfo(0, "swapper", 0, 0),
		NewInfo(100, "myapp", 1024, 0),
	}
	result := Find(infos, []string{"swapper", "myapp"}, -1, 0)
	for _, p := range result {
		assert.Greater(t, p.PID, 1, "filter should exclude PID %d", p.PID)
	}
	require.Len(t, result, 1)
	assert.Equal(t, 100, result[0].PID)
}
