//go:build integration

package pscommand

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestIntegrationNewManager_ReturnsProcessManagerWhenPsIsAvailable(t *testing.T) {
	// When
	processManager, err := NewManager()

	// Then
	require.NoError(t, err)
	assert.NotNil(t, processManager)
}

func TestIntegrationManager_Discover_ReturnsWellFormedProcesses(t *testing.T) {
	// Given
	processManager := newRealManager(t)

	// When
	processes, err := processManager.Discover()

	// Then
	require.NoError(t, err)
	require.NotEmpty(t, processes)
	for _, info := range processes {
		assert.Greater(t, info.PID, 0, "invalid PID")
		assert.NotEmpty(t, info.Name, "empty name for PID %d", info.PID)
		assert.GreaterOrEqual(t, info.Rss, int64(0), "negative RSS for PID %d", info.PID)
		assert.NotRegexp(t, "^/", info.Name, "PID %d has a full path in its name", info.PID)
	}
}

func TestIntegrationManager_OwnPID_ReturnsPIDOfCallingProcess(t *testing.T) {
	// Given
	processManager := newRealManager(t)
	expected := os.Getpid()

	// When
	actual := processManager.OwnPID()

	// Then
	assert.Equal(t, expected, actual)
}

func TestIntegrationManager_OwnUID_ReturnsEffectiveUIDOfCallingProcess(t *testing.T) {
	// Given
	processManager := newRealManager(t)
	expected := os.Geteuid()

	// When
	actual := processManager.OwnUID()

	// Then
	assert.Equal(t, expected, actual)
}

func TestIntegrationManager_LookupName_ReturnsNameOfRunningProcess(t *testing.T) {
	// Given
	processManager := newRealManager(t)

	// When
	name, err := processManager.LookupName(os.Getpid())

	// Then
	require.NoError(t, err)
	assert.NotEmpty(t, name)
}

func TestIntegrationManager_LookupName_ReturnsRealNameDespiteSpoofedArgv0(t *testing.T) {
	// Given
	sleepPath, err := exec.LookPath("sleep")
	require.NoError(t, err)
	command := &exec.Cmd{Path: sleepPath, Args: []string{"TOTALLY-DIFFERENT-SPOOFED-NAME", "30"}}
	require.NoError(t, command.Start())
	t.Cleanup(func() { _ = command.Process.Kill() })
	processManager := newRealManager(t)

	// When
	name, err := processManager.LookupName(command.Process.Pid)

	// Then
	require.NoError(t, err)
	assert.Equal(t, "sleep", name, "LookupName must report the real executable name, not a spoofed argv[0]")
}

func TestIntegrationManager_LookupName_ReturnsNotFoundErrorForExitedProcess(t *testing.T) {
	// Given
	pid := testutil.ExitedProcessPID(t)
	processManager := newRealManager(t)

	// When
	_, err := processManager.LookupName(pid)

	// Then
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
}

func TestIntegrationManager_Pin_ReturnsHandleForExistingProcess(t *testing.T) {
	// Given
	processManager := newRealManager(t)

	// When
	processHandle, err := processManager.Pin(os.Getpid())

	// Then
	require.NoError(t, err)
	assert.NotNil(t, processHandle)
}

func TestIntegrationManager_Pin_ReturnsNotFoundErrorForExitedProcess(t *testing.T) {
	// Given
	pid := testutil.ExitedProcessPID(t)
	processManager := newRealManager(t)

	// When
	processHandle, err := processManager.Pin(pid)

	// Then
	assert.Nil(t, processHandle)
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
}

func newRealManager(t *testing.T) outbound.ProcessManager {
	t.Helper()
	processManager, err := NewManager()
	require.NoError(t, err)

	return processManager
}
