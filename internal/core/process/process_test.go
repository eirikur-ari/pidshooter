package process

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_NoPatterns(t *testing.T) {
	require.ErrorIs(t, Validate(nil), ErrNoPatterns)
	require.ErrorIs(t, Validate([]string{}), ErrNoPatterns)
}

func TestValidate_TooShort(t *testing.T) {
	for _, p := range []string{"", "a", "ab"} {
		assert.Error(t, Validate([]string{p}), "expected error for pattern %q shorter than MinPatternLength", p)
	}
}

func TestValidate_ExactMinLength(t *testing.T) {
	p := strings.Repeat("a", MinPatternLength)
	assert.NoError(t, Validate([]string{p}), "expected no error for min-length pattern")
}

func TestValidate_Valid(t *testing.T) {
	assert.NoError(t, Validate([]string{"firefox"}))
}

func TestValidate_ExactMaxLength(t *testing.T) {
	p := strings.Repeat("a", MaxPatternLength)
	assert.NoError(t, Validate([]string{p}), "expected no error for max-length pattern")
}

func TestValidate_TooLong(t *testing.T) {
	p := strings.Repeat("a", MaxPatternLength+1)
	assert.Error(t, Validate([]string{p}), "expected error for pattern exceeding MaxPatternLength")
}

func TestErrNoPatterns_IsSentinel(t *testing.T) {
	require.ErrorIs(t, ErrNoPatterns, ErrNoPatterns, "ErrNoPatterns must satisfy errors.Is against itself")
}

func TestInfo_IsProtected(t *testing.T) {
	assert.True(t, NewInfo(0, "swapper", 0).IsProtected())
	assert.True(t, NewInfo(1, "init", 0).IsProtected())
	assert.False(t, NewInfo(2, "init", 0).IsProtected())
	assert.False(t, NewInfo(100, "myapp", 0).IsProtected())
}

func TestFind_MatchesByName(t *testing.T) {
	infos := []Info{
		NewInfo(100, "myapp", 1024),
		NewInfo(200, "worker", 2048),
	}
	result := Find(infos, []string{"myapp"}, 0)
	require.Len(t, result, 1)
	assert.Equal(t, 100, result[0].Pid)
}

func TestFind_SubstringMatch(t *testing.T) {
	infos := []Info{NewInfo(100, "myapp-worker", 1024)}
	result := Find(infos, []string{"app"}, 0)
	assert.Len(t, result, 1)
}

func TestFind_CaseInsensitive(t *testing.T) {
	infos := []Info{NewInfo(100, "MyApp", 1024)}
	result := Find(infos, []string{"myapp"}, 0)
	assert.Len(t, result, 1)
}

func TestFind_MultipleTerms(t *testing.T) {
	infos := []Info{
		NewInfo(100, "myapp", 1024),
		NewInfo(200, "worker", 2048),
		NewInfo(300, "other", 512),
	}
	result := Find(infos, []string{"myapp", "worker"}, 0)
	assert.Len(t, result, 2)
}

func TestFind_NoMatch(t *testing.T) {
	infos := []Info{NewInfo(100, "myapp", 1024)}
	result := Find(infos, []string{"worker"}, 0)
	assert.Empty(t, result)
}

func TestFind_ExcludesPID1(t *testing.T) {
	infos := []Info{
		NewInfo(1, "init", 512),
		NewInfo(100, "myapp", 1024),
	}
	result := Find(infos, []string{"init", "myapp"}, 0)
	for _, p := range result {
		assert.NotEqual(t, 1, p.Pid, "should not include PID 1")
	}
}

func TestFind_ExcludesOwnPID(t *testing.T) {
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

func TestFind_ExcludesPID0(t *testing.T) {
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
