//go:build integration

package osprocess_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/osprocess"
)

func TestIntegrationDiscoverReturnsResults(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	processes, err := f.Discover()
	require.NoError(t, err)
	assert.NotEmpty(t, processes)
}

func TestIntegrationDiscoverValidFields(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	processes, err := f.Discover()
	require.NoError(t, err)
	for _, p := range processes {
		assert.Greater(t, p.PID, 0, "invalid PID")
		assert.NotEmpty(t, p.Name, "empty name for PID %d", p.PID)
		assert.GreaterOrEqual(t, p.Rss, int64(0), "negative RSS for PID %d", p.PID)
	}
}

func TestIntegrationDiscoverShortProcessNames(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	processes, err := f.Discover()
	require.NoError(t, err)
	for _, p := range processes {
		assert.False(t, strings.HasPrefix(p.Name, "/"), "PID %d has a full path in name: %q", p.PID, p.Name)
	}
}

func TestIntegrationFindExcludesOwnAndInitPID(t *testing.T) {
	f, err := osprocess.NewProcess()
	require.NoError(t, err)
	raw, err := f.Discover()
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

func TestIntegrationLookupNameResistsArgv0Spoofing(t *testing.T) {
	sleepPath, err := exec.LookPath("sleep")
	require.NoError(t, err)

	cmd := &exec.Cmd{Path: sleepPath, Args: []string{"TOTALLY-DIFFERENT-SPOOFED-NAME", "30"}}
	require.NoError(t, cmd.Start())
	defer func() { _ = cmd.Process.Kill() }()

	f, err := osprocess.NewProcess()
	require.NoError(t, err)

	name, err := f.LookupName(cmd.Process.Pid)
	require.NoError(t, err)
	assert.Equal(t, "sleep", name, "LookupName must report the real executable name, not a spoofed argv[0]")
}
