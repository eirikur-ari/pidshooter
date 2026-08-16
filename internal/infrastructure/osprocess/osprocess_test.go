package osprocess

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestProcess(t *testing.T) *Process {
	t.Helper()
	path, err := exec.LookPath("ps")
	require.NoError(t, err, "ps not found")
	return &Process{psPath: path}
}

func TestListReturnsResults(t *testing.T) {
	processes, err := newTestProcess(t).List()
	require.NoError(t, err)
	assert.NotEmpty(t, processes)
}

func TestOwnPidMatchesOSGetpid(t *testing.T) {
	p := newTestProcess(t)
	assert.Equal(t, os.Getpid(), p.OwnPid())
}

func TestLookupNameReturnsOwnName(t *testing.T) {
	p := newTestProcess(t)
	name, err := p.LookupName(os.Getpid())
	require.NoError(t, err)
	assert.NotEmpty(t, name)
}

// Kill no longer guards pid or verifies name itself (see the Killer doc
// comment on outbound.Process) — that's now service.kill's job, using
// Info.IsProtected and LookupName before ever calling Kill. Exercising
// Kill against pid 0/1/-1/self here would send a real SIGKILL to the
// process group, init, or the test binary itself, so this only checks the
// one safe, deterministic case: a PID that doesn't exist.
func TestKillerNonexistentPID(t *testing.T) {
	_, err := newTestProcess(t).Kill(999999)
	assert.Error(t, err)
}
