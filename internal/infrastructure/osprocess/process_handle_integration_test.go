//go:build integration

package osprocess_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/osprocess"
)

func TestIntegrationPinThenKillReportsNotFoundAfterProcessExits(t *testing.T) {
	cmd := exec.Command("true")
	require.NoError(t, cmd.Start())

	f, err := osprocess.NewProcess()
	require.NoError(t, err)

	handle, err := f.Pin(cmd.Process.Pid)
	require.NoError(t, err)

	require.NoError(t, cmd.Wait(), "the process must actually exit and be reaped before Kill is attempted")

	err = handle.Kill()
	require.Error(t, err)
	assert.ErrorAs(t, err, &outbound.NotFoundError{}, "killing a handle pinned before the process exited must still report NotFoundError, not silently succeed")
}