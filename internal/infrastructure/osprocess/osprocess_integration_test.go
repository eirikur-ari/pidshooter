//go:build integration

package osprocess_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/infrastructure/osprocess"
)

func TestIntegration_List_ReturnsResults(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	processes, err := f.List()
	require.NoError(t, err)
	assert.NotEmpty(t, processes)
}

func TestIntegration_List_ValidFields(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	processes, err := f.List()
	require.NoError(t, err)
	for _, p := range processes {
		assert.Greater(t, p.Pid, 0, "invalid PID")
		assert.NotEmpty(t, p.Name, "empty name for PID %d", p.Pid)
		assert.GreaterOrEqual(t, p.Rss, int64(0), "negative RSS for PID %d", p.Pid)
	}
}

func TestIntegration_List_ShortProcessNames(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	processes, err := f.List()
	require.NoError(t, err)
	for _, p := range processes {
		assert.NotContains(t, p.Name, "/", "PID %d has a full path in name: %q", p.Pid, p.Name)
	}
}

func TestIntegration_Find_ExcludesOwnPID(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	ownPID := os.Getpid()
	results, err := f.Find([]string{"proc"})
	require.NoError(t, err)
	for _, p := range results {
		assert.NotEqual(t, ownPID, p.Pid, "own PID should be excluded")
		assert.NotEqual(t, 1, p.Pid, "PID 1 should be excluded")
	}
}
