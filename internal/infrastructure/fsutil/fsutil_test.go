package fsutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDirJoinsHomeAndAppName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir, err := ConfigDir("pidshooter")

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "pidshooter"), dir)
}

func TestConfigDirReturnsErrorWhenHomeUnset(t *testing.T) {
	t.Setenv("HOME", "")

	_, err := ConfigDir("pidshooter")

	assert.Error(t, err)
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

	dir := t.TempDir()
	path := filepath.Join(dir, "scores.json")
	require.NoError(t, WriteFileAtomic(path, []byte("original"), 0600))

	require.NoError(t, os.Chmod(dir, 0500))
	err := WriteFileAtomic(path, []byte("replacement"), 0600)
	require.NoError(t, os.Chmod(dir, 0700))

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
