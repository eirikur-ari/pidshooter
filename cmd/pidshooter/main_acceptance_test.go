//go:build acceptance

package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestAcceptance_Help_SucceedsWithoutTerm(t *testing.T) {
	// Given
	binary := pidshooterBinary(t)
	testutil.UnsetEnv(t, "TERM")
	cmd := exec.Command(binary, "--help")

	// When
	out, err := cmd.CombinedOutput()

	// Then
	require.NoError(t, err, "expected --help to succeed without $TERM set, got: %s", out)
	assert.Contains(t, string(out), "Process ID Shooter")
}

func TestAcceptance_UnknownFlag_ExitsWithCodeTwoAndPrintsUsageToStderr(t *testing.T) {
	// Given
	binary := pidshooterBinary(t)
	cmd := exec.Command(binary, "play", "--unknown")
	var exitErr *exec.ExitError
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// When
	err := cmd.Run()

	// Then
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 2, exitErr.ExitCode())
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "unknown flag: --unknown")
	assert.Contains(t, stderr.String(), "Usage:")
}

func TestAcceptance_UnknownCommand_ExitsWithCodeTwoAndPrintsUsageToStderr(t *testing.T) {
	// Given
	binary := pidshooterBinary(t)
	cmd := exec.Command(binary, "chrome")
	var exitErr *exec.ExitError
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// When
	err := cmd.Run()

	// Then
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 2, exitErr.ExitCode())
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "unknown command: chrome")
	assert.Contains(t, stderr.String(), "Process ID Shooter")
	assert.Contains(t, stderr.String(), "Usage:")
}

func pidshooterBinary(t *testing.T) string {
	t.Helper()
	binary := os.Getenv("PIDSHOOTER_BINARY")
	require.NotEmpty(t, binary, "PIDSHOOTER_BINARY must point to the built binary; run via make test-acceptance")
	return binary
}
