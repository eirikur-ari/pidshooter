package osprocess

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func newTestProcess(t *testing.T) *Process {
	t.Helper()
	path, err := exec.LookPath("ps")
	require.NoError(t, err, "ps not found")
	return &Process{psPath: path}
}

func TestValidate_EmptySlice(t *testing.T) {
	err := validate([]string{})
	require.Error(t, err)
	assert.ErrorIs(t, err, process.ErrNoPatterns)
}

func TestValidate_EmptyTerm(t *testing.T) {
	assert.Error(t, validate([]string{""}))
}

func TestValidate_TooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		assert.Error(t, validate([]string{p}), "expected error for pattern %q shorter than MinPatternLength", p)
	}
}

func TestValidate_ExactMinLength(t *testing.T) {
	min := strings.Repeat("a", process.MinPatternLength)
	assert.NoError(t, validate([]string{min}))
}

func TestValidate_TooLong(t *testing.T) {
	long := strings.Repeat("a", process.MaxPatternLength+1)
	assert.Error(t, validate([]string{long}))
}

func TestValidate_Valid(t *testing.T) {
	assert.NoError(t, validate([]string{"myapp", "worker"}))
}

func TestValidate_ExactMaxLength(t *testing.T) {
	exact := strings.Repeat("a", process.MaxPatternLength)
	assert.NoError(t, validate([]string{exact}))
}

func TestFind_EmptyPatterns(t *testing.T) {
	_, err := newTestProcess(t).Find([]string{})
	assert.Error(t, err)
}

func TestFind_EmptyTerm(t *testing.T) {
	_, err := newTestProcess(t).Find([]string{""})
	assert.Error(t, err)
}

func TestFilter_MatchesByName(t *testing.T) {
	processes := []process.Info{
		{Pid: 100, Name: "myapp", Rss: 1024},
		{Pid: 200, Name: "worker", Rss: 2048},
	}
	result := filter(processes, []string{"myapp"})
	require.Len(t, result, 1)
	assert.Equal(t, 100, result[0].Pid)
}

func TestFilter_SubstringMatch(t *testing.T) {
	processes := []process.Info{
		{Pid: 100, Name: "myapp-worker", Rss: 1024},
	}
	result := filter(processes, []string{"app"})
	assert.Len(t, result, 1)
}

func TestFilter_CaseInsensitive(t *testing.T) {
	processes := []process.Info{
		{Pid: 100, Name: "MyApp", Rss: 1024},
	}
	result := filter(processes, []string{"myapp"})
	assert.Len(t, result, 1)
}

func TestFilter_MultipleTerms(t *testing.T) {
	processes := []process.Info{
		{Pid: 100, Name: "myapp", Rss: 1024},
		{Pid: 200, Name: "worker", Rss: 2048},
		{Pid: 300, Name: "other", Rss: 512},
	}
	result := filter(processes, []string{"myapp", "worker"})
	assert.Len(t, result, 2)
}

func TestFilter_NoMatch(t *testing.T) {
	processes := []process.Info{
		{Pid: 100, Name: "myapp", Rss: 1024},
	}
	result := filter(processes, []string{"worker"})
	assert.Empty(t, result)
}

func TestFilter_ExcludesPID1(t *testing.T) {
	processes := []process.Info{
		{Pid: 1, Name: "init", Rss: 512},
		{Pid: 100, Name: "myapp", Rss: 1024},
	}
	result := filter(processes, []string{"init", "myapp"})
	for _, p := range result {
		assert.NotEqual(t, 1, p.Pid, "should not include PID 1")
	}
}

func TestFilter_ExcludesOwnPID(t *testing.T) {
	ownPID := os.Getpid()
	processes := []process.Info{
		{Pid: ownPID, Name: "testprocess", Rss: 1024},
		{Pid: 100, Name: "testprocess", Rss: 2048},
	}
	result := filter(processes, []string{"testprocess"})
	for _, p := range result {
		assert.NotEqual(t, ownPID, p.Pid, "should not include own PID")
	}
	require.Len(t, result, 1)
	assert.Equal(t, 100, result[0].Pid)
}

func TestFilter_ExcludesPID0(t *testing.T) {
	processes := []process.Info{
		{Pid: 0, Name: "swapper", Rss: 0},
		{Pid: 100, Name: "myapp", Rss: 1024},
	}
	result := filter(processes, []string{"swapper", "myapp"})
	for _, p := range result {
		assert.Greater(t, p.Pid, 1, "filter should exclude PID %d", p.Pid)
	}
	require.Len(t, result, 1)
	assert.Equal(t, 100, result[0].Pid)
}

func TestKiller_InvalidPID(t *testing.T) {
	_, err := newTestProcess(t).Kill(-1, "")
	assert.Error(t, err)
}

func TestKiller_RefusesPID0(t *testing.T) {
	_, err := newTestProcess(t).Kill(0, "")
	assert.Error(t, err)
}

func TestKiller_RefusesPID1(t *testing.T) {
	_, err := newTestProcess(t).Kill(1, "")
	assert.Error(t, err)
}

func TestKiller_RefusesNegativePID(t *testing.T) {
	p := newTestProcess(t)
	for _, pid := range []int{-1, -100, -99999} {
		_, err := p.Kill(pid, "")
		assert.Error(t, err, "expected error killing PID %d", pid)
	}
}

func TestValidateProcessName_Match(t *testing.T) {
	assert.NoError(t, validateProcessName("myapp", "myapp"))
}

func TestValidateProcessName_Mismatch(t *testing.T) {
	assert.Error(t, validateProcessName("myapp", "otherd"))
}

func TestValidateProcessName_EmptyActual(t *testing.T) {
	assert.Error(t, validateProcessName("myapp", ""))
}

func TestKiller_RefusesNameMismatch(t *testing.T) {
	ownPID := os.Getpid()
	killed, err := newTestProcess(t).Kill(ownPID, "definitely-not-this-process")
	assert.NoError(t, err)
	assert.False(t, killed, "expected not killed when name does not match current process name")
}
