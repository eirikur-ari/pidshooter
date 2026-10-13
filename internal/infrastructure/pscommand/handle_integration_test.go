//go:build integration

package pscommand

import (
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestIntegrationHandle_Kill_TerminatesProcess(t *testing.T) {
	// Given
	command := exec.Command("sleep", "30")
	require.NoError(t, command.Start())
	t.Cleanup(func() { _ = command.Process.Kill() })
	processHandle, err := newRealManager(t).Pin(command.Process.Pid)
	require.NoError(t, err)

	// When
	killErr := processHandle.Kill()
	waitErr := command.Wait()
	signaled := command.ProcessState.Sys().(syscall.WaitStatus).Signaled()

	// Then
	require.NoError(t, killErr)
	assert.Error(t, waitErr)
	assert.True(t, signaled, "the process must have been terminated by a signal")
}

func TestIntegrationHandle_Kill_ReturnsNotFoundErrorAfterProcessExits(t *testing.T) {
	// Given
	command := exec.Command("/usr/bin/true")
	require.NoError(t, command.Start())
	processHandle, err := newRealManager(t).Pin(command.Process.Pid)
	require.NoError(t, err)
	require.NoError(t, command.Wait(), "the process must actually exit and be reaped before Kill is attempted")

	// When
	err = processHandle.Kill()

	// Then
	assert.ErrorAs(t, err, &outbound.NotFoundError{}, "killing a handle pinned before the process exited must still report NotFoundError, not silently succeed")
}

func TestIntegrationHandle_Kill_ReturnsOtherSignalErrorsUnclassified(t *testing.T) {
	// Given
	processHandle, err := newRealManager(t).Pin(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, processHandle.Release())

	// When
	err = processHandle.Kill()

	// Then
	require.Error(t, err)
	assert.NotErrorAs(t, err, &outbound.NotFoundError{})
}

func TestIntegrationHandle_Release_SucceedsForExistingProcess(t *testing.T) {
	// Given
	processHandle, err := newRealManager(t).Pin(os.Getpid())
	require.NoError(t, err)

	// When
	err = processHandle.Release()

	// Then
	assert.NoError(t, err)
}
