//go:build integration

package osprocess_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/osprocess"
)

func TestIntegrationListReturnsResults(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	processes, err := f.List()
	require.NoError(t, err)
	assert.NotEmpty(t, processes)
}

func TestIntegrationListValidFields(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	processes, err := f.List()
	require.NoError(t, err)
	for _, p := range processes {
		assert.Greater(t, p.PID, 0, "invalid PID")
		assert.NotEmpty(t, p.Name, "empty name for PID %d", p.PID)
		assert.GreaterOrEqual(t, p.Rss, int64(0), "negative RSS for PID %d", p.PID)
	}
}

func TestIntegrationListShortProcessNames(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	processes, err := f.List()
	require.NoError(t, err)
	for _, p := range processes {
		assert.False(t, strings.HasPrefix(p.Name, "/"), "PID %d has a full path in name: %q", p.PID, p.Name)
	}
}

// PID exclusion (self and PID<=1) is domain policy applied by process.Find,
// not something the adapter does anymore. This exercises the real List/OwnPID
// adapter output through the real domain filter end-to-end.
func TestIntegrationFindExcludesOwnAndInitPID(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	raw, err := f.List()
	require.NoError(t, err)

	infos := make([]process.Info, len(raw))
	for i, p := range raw {
		infos[i] = process.NewInfo(p.PID, p.Name, p.Rss, p.UID)
	}

	ownPID := f.OwnPID()
	ownUID := f.OwnUID()
	result := process.Find(infos, []string{"proc"}, ownPID, ownUID)
	for _, p := range result {
		assert.NotEqual(t, ownPID, p.PID, "own PID should be excluded")
		assert.NotEqual(t, 1, p.PID, "PID 1 should be excluded")
		if ownUID != 0 {
			assert.Equal(t, ownUID, p.UID, "non-root caller should only see processes it owns")
		}
	}
}
