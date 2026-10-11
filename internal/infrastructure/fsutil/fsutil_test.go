package fsutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestConfigDir_JoinsHomeAndAppNameWhenXDGConfigHomeUnset(t *testing.T) {
	// Given
	home := t.TempDir()
	t.Setenv("HOME", home)
	testutil.UnsetEnv(t, "XDG_CONFIG_HOME")
	expected := filepath.Join(home, ".config", "pidshooter")

	// When
	dir, err := ConfigDir("pidshooter")

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, dir)
}

func TestConfigDir_PrefersAbsoluteXDGConfigHome(t *testing.T) {
	// Given
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	expected := filepath.Join(xdg, "pidshooter")

	// When
	dir, err := ConfigDir("pidshooter")

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, dir)
}

func TestConfigDir_IgnoresRelativeXDGConfigHome(t *testing.T) {
	// Given
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "relative/path")
	expected := filepath.Join(home, ".config", "pidshooter")

	// When
	dir, err := ConfigDir("pidshooter")

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, dir)
}

func TestConfigDir_ReturnsErrorWhenHomeEmpty(t *testing.T) {
	// Given
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	// When
	_, err := ConfigDir("pidshooter")

	// Then
	assert.Error(t, err)
}

func TestWriteFileAtomic_CreatesMissingParentDirectories(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "nested", "deeper", "scores.json")
	expected := []byte("hello")

	// When
	err := WriteFileAtomic(path, expected, 0600)
	data, readErr := os.ReadFile(path)

	// Then
	require.NoError(t, err)
	require.NoError(t, readErr)
	assert.Equal(t, expected, data)
}

func TestWriteFileAtomic_SetsRequestedPermissions(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "scores.json")
	expected := os.FileMode(0600)

	// When
	err := WriteFileAtomic(path, []byte("hello"), expected)
	info, statErr := os.Stat(path)

	// Then
	require.NoError(t, err)
	require.NoError(t, statErr)
	assert.Equal(t, expected, info.Mode().Perm())
}

func TestWriteFileAtomic_ReplacesExistingFile(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "scores.json")
	require.NoError(t, WriteFileAtomic(path, []byte("first"), 0600))
	expected := []byte("second")

	// When
	err := WriteFileAtomic(path, expected, 0600)
	data, readErr := os.ReadFile(path)

	// Then
	require.NoError(t, err)
	require.NoError(t, readErr)
	assert.Equal(t, expected, data)
}

func TestWriteFileAtomic_LeavesNoTempFileAfterSuccess(t *testing.T) {
	// Given
	dir := t.TempDir()
	path := filepath.Join(dir, "scores.json")

	// When
	err := WriteFileAtomic(path, []byte("hello"), 0600)
	leftovers := tempFilesIn(t, dir)

	// Then
	require.NoError(t, err)
	assert.Empty(t, leftovers)
}

func TestWriteFileAtomic_LeavesExistingFileUntouchedOnFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses permission bits, making this test meaningless")
	}

	// Given
	// The failure is forced via the parent because dir's own permissions are
	// reset to 0700 on every call. Removing the parent's execute bit blocks
	// traversal into dir before WriteFileAtomic can reset anything.
	parent := t.TempDir()
	dir := filepath.Join(parent, "cfgdir")
	path := filepath.Join(dir, "scores.json")
	expected := []byte("original")
	require.NoError(t, WriteFileAtomic(path, expected, 0600))
	t.Cleanup(func() { _ = os.Chmod(parent, 0700) })
	require.NoError(t, os.Chmod(parent, 0600))

	// When
	err := WriteFileAtomic(path, []byte("replacement"), 0600)
	restoreErr := os.Chmod(parent, 0700)
	data, readErr := os.ReadFile(path)
	leftovers := tempFilesIn(t, dir)

	// Then
	assert.Error(t, err)
	require.NoError(t, restoreErr)
	require.NoError(t, readErr)
	assert.Equal(t, expected, data, "a failed write must not corrupt the existing file")
	assert.Empty(t, leftovers, "no temp file should remain after a failed write")
}

func TestWriteFileAtomic_ReturnsErrorAndLeavesNoTempFileWhenPathIsADirectory(t *testing.T) {
	// Given
	dir := t.TempDir()
	path := filepath.Join(dir, "scores.json")
	require.NoError(t, os.Mkdir(path, 0700))

	// When
	err := WriteFileAtomic(path, []byte("hello"), 0600)
	info, statErr := os.Stat(path)
	leftovers := tempFilesIn(t, dir)

	// Then
	assert.Error(t, err)
	require.NoError(t, statErr)
	assert.True(t, info.IsDir(), "the directory at path must not be replaced")
	assert.Empty(t, leftovers, "no temp file should remain after a failed rename")
}

func TestWriteFileAtomic_TightensExistingDirectoryPermissions(t *testing.T) {
	// Given
	dir := filepath.Join(t.TempDir(), "cfgdir")
	require.NoError(t, os.Mkdir(dir, 0777))
	path := filepath.Join(dir, "scores.json")
	expected := os.FileMode(0700)

	// When
	err := WriteFileAtomic(path, []byte("hello"), 0600)
	info, statErr := os.Stat(dir)

	// Then
	require.NoError(t, err)
	require.NoError(t, statErr)
	assert.Equal(t, expected, info.Mode().Perm())
}

func tempFilesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	var names []string
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp-") {
			names = append(names, entry.Name())
		}
	}
	return names
}
