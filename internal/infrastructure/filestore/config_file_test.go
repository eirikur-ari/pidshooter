package filestore

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestNewConfigFile_ResolvesDefaultPath(t *testing.T) {
	// Given
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	expectedPath := filepath.Join(home, ".config", "pidshooter", "config.yaml")

	// When
	store, err := NewConfigFile()
	typed, ok := store.(*configFile)

	// Then
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, expectedPath, typed.file.path)
}

func TestNewConfigFile_ReturnsErrorWhenHomeUnset(t *testing.T) {
	// Given
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	// When
	store, err := NewConfigFile()

	// Then
	assert.Error(t, err)
	assert.Nil(t, store)
}

func TestConfigFile_Load_ReturnsNotFoundErrorWhenFileMissing(t *testing.T) {
	// Given
	store := newConfigFileFixture(t.TempDir())

	// When
	config, err := store.Load()

	// Then
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
	assert.Empty(t, config)
}

func TestConfigFile_Load_ReturnsCorruptedDataErrorForUndecodableContent(t *testing.T) {
	tests := newCorruptedConfigTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			store := newConfigFileFixture(t.TempDir())
			require.NoError(t, os.WriteFile(store.file.path, []byte(test.content), 0600))

			// When
			_, err := store.Load()

			// Then
			assert.ErrorAs(t, err, &outbound.CorruptedDataError{})
		})
	}
}

func TestConfigFile_Load_ReturnsCorruptedDataErrorWhenFileOverMaximumSize(t *testing.T) {
	// Given
	store := newConfigFileFixture(t.TempDir())
	require.NoError(t, os.WriteFile(store.file.path, make([]byte, maxConfigFileSize+1), 0600))

	// When
	_, err := store.Load()

	// Then
	assert.ErrorAs(t, err, &outbound.CorruptedDataError{})
}

func TestConfigFile_Load_ReturnsEmptyConfigForEmptyFile(t *testing.T) {
	// Given
	store := newConfigFileFixture(t.TempDir())
	require.NoError(t, os.WriteFile(store.file.path, nil, 0600))

	// When
	config, err := store.Load()

	// Then
	require.NoError(t, err)
	assert.Empty(t, config)
}

func TestConfigFile_Load_AcceptsFileWithoutVersion(t *testing.T) {
	// Given
	store := newConfigFileFixture(t.TempDir())
	require.NoError(t, os.WriteFile(store.file.path, []byte("mode: lucky\n"), 0600))

	// When
	config, err := store.Load()

	// Then
	require.NoError(t, err)
	assert.Equal(t, outbound.ModeLucky, config.Mode)
}

func TestConfigFile_Load_RejectsNewerSchemaVersionWithoutClassifyingIt(t *testing.T) {
	// Given
	store := newConfigFileFixture(t.TempDir())
	require.NoError(t, os.WriteFile(store.file.path, []byte("version: 999\n"), 0600))

	// When
	_, err := store.Load()

	// Then
	require.Error(t, err)
	assert.NotErrorAs(t, err, &outbound.NotFoundError{})
	assert.NotErrorAs(t, err, &outbound.CorruptedDataError{})
}

func TestConfigFile_Load_PassesThroughUnrecognizedMode(t *testing.T) {
	// Given
	store := newConfigFileFixture(t.TempDir())
	require.NoError(t, os.WriteFile(store.file.path, []byte("mode: wobble\n"), 0600))
	expectedMode := outbound.Mode("wobble")

	// When
	config, err := store.Load()

	// Then
	require.NoError(t, err)
	assert.Equal(t, expectedMode, config.Mode)
}

func TestConfigFile_Load_LeavesAbsentKeysNil(t *testing.T) {
	// Given
	store := newConfigFileFixture(t.TempDir())
	require.NoError(t, os.WriteFile(store.file.path, []byte("game:\n  speed: 3.0\n"), 0600))

	// When
	config, err := store.Load()

	// Then
	require.NoError(t, err)
	require.NotNil(t, config.Game.Speed)
	assert.Equal(t, 3.0, *config.Game.Speed)
	assert.Nil(t, config.Game.ConfirmMode)
	assert.Nil(t, config.Game.TimeLimit)
	assert.Nil(t, config.Process.IncludeRoot)
}

func TestConfigFile_Save_ReturnsErrorWhenConfigCannotBeEncoded(t *testing.T) {
	// Given
	store := newConfigFileFixture(t.TempDir())
	encodeErr := errors.New("cannot encode")
	encoder := &MockFileEncoder{}
	encoder.On("encodeYAML", newConfigContentFixture()).Return(nil, encodeErr)
	store.encoder = encoder

	// When
	err := store.Save(newConfigFixture())
	_, statErr := os.Stat(store.file.path)

	// Then
	assert.ErrorIs(t, err, encodeErr)
	assert.ErrorIs(t, statErr, fs.ErrNotExist)
	encoder.AssertExpectations(t)
}

func TestConfigFile_Save_ReturnsErrorWhenFileCannotBeWritten(t *testing.T) {
	// Given
	parent := filepath.Join(t.TempDir(), "parent")
	require.NoError(t, os.WriteFile(parent, nil, 0600))
	store := newConfigFileFixture(t.TempDir())
	store.file.path = filepath.Join(parent, "config.yaml")

	// When
	err := store.Save(newConfigFixture())

	// Then
	assert.Error(t, err)
}

func TestConfigFile_SaveLoad_RoundTripsConfig(t *testing.T) {
	// Given
	store := newConfigFileFixture(t.TempDir())
	config := newConfigFixture()

	// When
	saveErr := store.Save(config)
	loaded, loadErr := store.Load()

	// Then
	require.NoError(t, saveErr)
	require.NoError(t, loadErr)
	assert.Equal(t, config, loaded)
}

func TestConfigFile_Save_WritesYAMLMatchingOnDiskSchema(t *testing.T) {
	tests := newSavedYAMLTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			store := newConfigFileFixture(t.TempDir())
			var onDisk map[string]any

			// When
			saveErr := store.Save(test.config)
			raw, readErr := os.ReadFile(store.file.path)
			unmarshalErr := yaml.Unmarshal(raw, &onDisk)

			// Then
			require.NoError(t, saveErr)
			require.NoError(t, readErr)
			require.NoError(t, unmarshalErr)
			assert.Equal(t, test.expected, onDisk)
		})
	}
}

func newSavedYAMLTestCases() []struct {
	name     string
	config   outbound.Config
	expected map[string]any
} {
	return []struct {
		name     string
		config   outbound.Config
		expected map[string]any
	}{
		{
			"every key set",
			newConfigFixture(),
			map[string]any{
				"version": currentConfigSchemaVersion,
				"mode":    "game",
				"process": map[string]any{"include_root": true},
				"game":    map[string]any{"confirm_mode": true, "speed": 2.5, "time_limit": 60},
			},
		},
		{
			"unset keys are omitted",
			outbound.Config{Game: outbound.GameConfig{Speed: testutil.Pointer(1.5)}},
			map[string]any{
				"version": currentConfigSchemaVersion,
				"mode":    "",
				"process": map[string]any{},
				"game":    map[string]any{"speed": 1.5},
			},
		},
	}
}

func newCorruptedConfigTestCases() []struct {
	name    string
	content string
} {
	return []struct {
		name    string
		content string
	}{
		{"invalid yaml", ":\n  - not: valid: yaml"},
		{"unknown nested key", "game:\n  timelimit: 55\n"},
		{"unknown top-level key", "bogus: true\n"},
	}
}
