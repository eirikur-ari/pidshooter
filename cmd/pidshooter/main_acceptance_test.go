//go:build acceptance

package main_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestAcceptance_Help_SucceedsWithoutTerm(t *testing.T) {
	// Given
	binary := buildPidshooter(t)
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
	binary := buildPidshooter(t)
	cmd := exec.Command(binary, "--unknown")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// When
	err := cmd.Run()

	// Then
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 2, exitErr.ExitCode())
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "flag provided but not defined: -unknown")
	assert.Contains(t, stderr.String(), "Process ID Shooter")
}

func buildPidshooter(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "pidshooter")
	build := exec.Command("go", "build", "-o", binary, ".")
	out, err := build.CombinedOutput()
	require.NoError(t, err, "go build failed: %s", out)
	return binary
}
