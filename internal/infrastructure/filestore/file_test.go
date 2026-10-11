package filestore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestNewFile_ResolvesPathUnderConfigDirectory(t *testing.T) {
	// Given
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	expectedPath := filepath.Join(home, ".config", "pidshooter", "data.txt")

	// When
	f, err := newFile("data.txt", 16)

	// Then
	require.NoError(t, err)
	assert.Equal(t, expectedPath, f.path)
	assert.Equal(t, 16, f.maxSize)
}

func TestNewFile_ReturnsErrorWhenHomeUnset(t *testing.T) {
	// Given
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	// When
	_, err := newFile("data.txt", 16)

	// Then
	assert.Error(t, err)
}

func TestFile_read_ReturnsContent(t *testing.T) {
	// Given
	f := newFileFixture(t.TempDir(), 16)
	require.NoError(t, os.WriteFile(f.path, []byte("content"), 0600))

	// When
	data, err := f.read()

	// Then
	require.NoError(t, err)
	assert.Equal(t, []byte("content"), data)
}

func TestFile_read_AcceptsContentOfExactlyMaximumSize(t *testing.T) {
	// Given
	f := newFileFixture(t.TempDir(), 16)
	require.NoError(t, os.WriteFile(f.path, make([]byte, 16), 0600))

	// When
	data, err := f.read()

	// Then
	require.NoError(t, err)
	assert.Len(t, data, 16)
}

func TestFile_read_ReturnsNotFoundErrorWhenFileMissing(t *testing.T) {
	// Given
	f := newFileFixture(t.TempDir(), 16)

	// When
	_, err := f.read()

	// Then
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
}

func TestFile_read_ReturnsCorruptedDataErrorWhenOverMaximumSize(t *testing.T) {
	// Given
	f := newFileFixture(t.TempDir(), 16)
	require.NoError(t, os.WriteFile(f.path, make([]byte, 17), 0600))

	// When
	_, err := f.read()

	// Then
	assert.ErrorAs(t, err, &outbound.CorruptedDataError{})
}

func TestFile_read_ReturnsOtherFailuresUnclassified(t *testing.T) {
	// Given
	f := file{path: t.TempDir(), maxSize: 16}

	// When
	_, err := f.read()

	// Then
	require.Error(t, err)
	assert.NotErrorAs(t, err, &outbound.NotFoundError{})
	assert.NotErrorAs(t, err, &outbound.CorruptedDataError{})
}

func TestFile_write_PersistsDataReadableByOwnerOnly(t *testing.T) {
	// Given
	f := newFileFixture(t.TempDir(), 16)

	expectedData := []byte("content")
	expectedPermissions := os.FileMode(0600)

	// When
	writeErr := f.write(expectedData)
	data, readErr := os.ReadFile(f.path)
	info, statErr := os.Stat(f.path)

	// Then
	require.NoError(t, writeErr)
	require.NoError(t, readErr)
	require.NoError(t, statErr)
	assert.Equal(t, expectedData, data)
	assert.Equal(t, expectedPermissions, info.Mode().Perm())
}

func TestFile_write_ReturnsErrorWhenParentIsNotADirectory(t *testing.T) {
	// Given
	parent := filepath.Join(t.TempDir(), "parent")
	require.NoError(t, os.WriteFile(parent, nil, 0600))
	f := newFileFixture(parent, 16)

	// When
	err := f.write([]byte("content"))

	// Then
	assert.Error(t, err)
}
