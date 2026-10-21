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

	// When
	cmd := exec.Command(binary, "--help")
	out, err := cmd.CombinedOutput()

	// Then
	require.NoError(t, err, "expected --help to succeed without $TERM set, got: %s", out)
	assert.Contains(t, string(out), "Process ID Shooter")
}

func TestAcceptance_UnknownFlag_ExitsWithCodeTwoAndPrintsUsageToStderr(t *testing.T) {
	// Given
	var exitErr *exec.ExitError
	var stdout, stderr bytes.Buffer
	binary := pidshooterBinary(t)

	// When
	cmd := exec.Command(binary, "play", "--unknown")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Then
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 2, exitErr.ExitCode())
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "unknown flag: --unknown")
	assert.Contains(t, stderr.String(), "Usage:")
}

func TestAcceptance_PlayWithoutPatterns_ExitsWithCodeTwoAndPrintsUsageToStderrWithoutWritingConfig(t *testing.T) {
	// Given
	var exitErr *exec.ExitError
	var stdout, stderr bytes.Buffer
	binary := pidshooterBinary(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	testutil.UnsetEnv(t, "XDG_CONFIG_HOME")

	// When
	cmd := exec.Command(binary, "play")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	entries, readErr := os.ReadDir(home)

	// Then
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 2, exitErr.ExitCode())
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "requires at least 1 arg")
	assert.Contains(t, stderr.String(), "Usage:")
	require.NoError(t, readErr)
	assert.Empty(t, entries)
}

func TestAcceptance_UnknownCommand_ExitsWithCodeTwoAndPrintsUsageToStderr(t *testing.T) {
	// Given
	var exitErr *exec.ExitError
	var stdout, stderr bytes.Buffer
	binary := pidshooterBinary(t)

	// When
	cmd := exec.Command(binary, "chrome")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

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
