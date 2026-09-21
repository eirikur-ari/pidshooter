//go:build acceptance

package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcceptanceHelpSucceedsWithoutTerm(t *testing.T) {
	binary := buildPidshooter(t)

	cmd := exec.Command(binary, "--help")
	cmd.Env = envWithout(os.Environ(), "TERM")

	out, err := cmd.CombinedOutput()

	require.NoError(t, err, "expected --help to succeed without $TERM set, got: %s", out)
	assert.Contains(t, string(out), "Process ID Shooter")
}

func buildPidshooter(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "pidshooter")
	build := exec.Command("go", "build", "-o", binary, ".")
	out, err := build.CombinedOutput()
	require.NoError(t, err, "go build failed: %s", out)
	return binary
}

func envWithout(env []string, key string) []string {
	prefix := key + "="
	filtered := make([]string, 0, len(env))
	for _, kv := range env {
		if !strings.HasPrefix(kv, prefix) {
			filtered = append(filtered, kv)
		}
	}
	return filtered
}
