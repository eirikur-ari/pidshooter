package fsutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil/helper"
)

func TestConfigDirJoinsHomeAndAppName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	helper.UnsetEnv(t, "XDG_CONFIG_HOME")

	dir, err := ConfigDir("pidshooter")

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "pidshooter"), dir)
}

func TestConfigDirReturnsErrorWhenHomeUnset(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	_, err := ConfigDir("pidshooter")

	assert.Error(t, err)
}

func TestConfigDirPrefersXDGConfigHomeWhenSet(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	dir, err := ConfigDir("pidshooter")

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(xdg, "pidshooter"), dir)
}

func TestConfigDirIgnoresRelativeXDGConfigHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "relative/path")

	dir, err := ConfigDir("pidshooter")

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "pidshooter"), dir)
}

func TestWriteFileAtomicCreatesParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "deeper", "scores.json")

	require.NoError(t, WriteFileAtomic(path, []byte("hello"), 0600))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}

func TestWriteFileAtomicSetsPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")

	require.NoError(t, WriteFileAtomic(path, []byte("hello"), 0600))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestWriteFileAtomicOverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "scores.json")

	require.NoError(t, WriteFileAtomic(path, []byte("first"), 0600))
	require.NoError(t, WriteFileAtomic(path, []byte("second"), 0600))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "second", string(data))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.NotContains(t, e.Name(), ".tmp-", "no temp file should remain after a successful write")
	}
}

func TestWriteFileAtomicLeavesExistingFileUntouchedOnFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission bits are not enforced the same way on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses permission bits, making this test meaningless")
	}

	// The failure has to be forced via the parent, not dir itself: dir's
	// own permissions get unconditionally reset to 0700 on every call (see
	// TestWriteFileAtomicTightensExistingDirectoryPermissions), so a
	// restrictive mode on dir wouldn't survive long enough to block
	// CreateTemp. Removing the parent's execute bit blocks traversal into
	// dir entirely, which happens before WriteFileAtomic ever gets a
	// chance to fix anything.
	parent := t.TempDir()
	dir := filepath.Join(parent, "cfgdir")
	path := filepath.Join(dir, "scores.json")
	require.NoError(t, WriteFileAtomic(path, []byte("original"), 0600))

	require.NoError(t, os.Chmod(parent, 0600))
	err := WriteFileAtomic(path, []byte("replacement"), 0600)
	require.NoError(t, os.Chmod(parent, 0700))

	assert.Error(t, err)
	data, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, "original", string(data), "a failed write must not corrupt the existing file")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.NotContains(t, e.Name(), ".tmp-", "no temp file should remain after a failed write")
	}
}

func TestWriteFileAtomicTightensExistingDirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission bits are not enforced the same way on Windows")
	}

	parent := t.TempDir()
	dir := filepath.Join(parent, "cfgdir")
	require.NoError(t, os.Mkdir(dir, 0777))
	path := filepath.Join(dir, "scores.json")

	require.NoError(t, WriteFileAtomic(path, []byte("hello"), 0600))

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
}
