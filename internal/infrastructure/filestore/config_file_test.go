package filestore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestNewConfigFileResolvesDefaultPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "") // don't let the runner's own env override HOME here

	c, err := NewConfigFile()

	require.NoError(t, err)
	cf, ok := c.(*configFile)
	require.True(t, ok)
	assert.Equal(t, filepath.Join(home, ".config", "pidshooter", "config.yaml"), cf.file.path)
}

func TestNewConfigFileReturnsErrorWhenHomeUnset(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "") // don't let the runner's own env mask the unset HOME

	_, err := NewConfigFile()

	assert.Error(t, err)
}

func TestConfigLoadFileNotExistReturnsNotFoundError(t *testing.T) {
	c := newTempConfig(t)
	defaults, err := c.Load()
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
	assert.Empty(t, defaults)
}

func TestConfigLoadInvalidYAMLReturnsCorruptedDataError(t *testing.T) {
	c := newTempConfig(t)
	require.NoError(t, os.WriteFile(c.file.path, []byte(":\n  - not: valid: yaml"), 0644))

	_, err := c.Load()

	assert.ErrorAs(t, err, &outbound.CorruptedDataError{})
}

func TestConfigLoadUnknownKeyReturnsCorruptedDataError(t *testing.T) {
	c := newTempConfig(t)
	require.NoError(t, os.WriteFile(c.file.path, []byte("game:\n  timelimit: 55\n"), 0644))

	_, err := c.Load()

	assert.ErrorAs(t, err, &outbound.CorruptedDataError{})
}

func TestConfigLoadEmptyFileReturnsNoError(t *testing.T) {
	c := newTempConfig(t)
	require.NoError(t, os.WriteFile(c.file.path, []byte(""), 0644))

	defaults, err := c.Load()

	require.NoError(t, err)
	assert.Empty(t, defaults)
}

func TestConfigSaveCreatesFile(t *testing.T) {
	c := newTempConfig(t)
	require.NoError(t, c.Save(outbound.Config{}))
	_, err := os.Stat(c.file.path)
	assert.NoError(t, err, "expected file to be created after Save")
}

func TestConfigSaveFilePermissions(t *testing.T) {
	c := newTempConfig(t)
	require.NoError(t, c.Save(outbound.Config{}))
	info, err := os.Stat(c.file.path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestConfigSaveLoadRoundTrip(t *testing.T) {
	c := newTempConfig(t)
	defaults := outbound.Config{
		Mode:    outbound.ModeGame,
		Process: outbound.ProcessConfig{IncludeRoot: testutil.Pointer(true)},
		Game: outbound.GameConfig{
			ConfirmMode: testutil.Pointer(true),
			Speed:       testutil.Pointer(2.5),
			TimeLimit:   testutil.Pointer(60),
		},
	}

	require.NoError(t, c.Save(defaults))
	loaded, err := c.Load()
	require.NoError(t, err)
	assert.Equal(t, defaults, loaded)
}

func TestConfigSaveOverwritesPreviousFile(t *testing.T) {
	c := newTempConfig(t)

	first := outbound.Config{Mode: outbound.ModeGame, Game: outbound.GameConfig{Speed: testutil.Pointer(1.0)}}
	require.NoError(t, c.Save(first))

	second := outbound.Config{Mode: outbound.ModeLucky, Game: outbound.GameConfig{Speed: testutil.Pointer(2.0)}}
	require.NoError(t, c.Save(second))

	loaded, err := c.Load()
	require.NoError(t, err)
	assert.Equal(t, second, loaded)
}

func TestConfigSaveToNestedNonexistentDirectoryCreatesParentDirs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "deeper", "config.yaml")
	c := newConfigFileAt(path)

	require.NoError(t, c.Save(outbound.Config{}))

	_, err := os.Stat(path)
	assert.NoError(t, err, "Save should create the path's parent directories, not ~/.config/pidshooter")
}

func TestConfigSaveDoesNotLeaveTempFileAfterSuccess(t *testing.T) {
	dir := t.TempDir()
	c := newConfigFileAt(filepath.Join(dir, "config.yaml"))

	require.NoError(t, c.Save(outbound.Config{}))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.NotContains(t, e.Name(), ".tmp-", "no temp file should remain after a successful save")
	}
}

func TestConfigSaveWritesYAMLMatchingOnDiskSchema(t *testing.T) {
	c := newTempConfig(t)
	defaults := outbound.Config{
		Mode:    outbound.ModeGame,
		Process: outbound.ProcessConfig{IncludeRoot: testutil.Pointer(true)},
		Game: outbound.GameConfig{
			ConfirmMode: testutil.Pointer(true),
			Speed:       testutil.Pointer(2.5),
			TimeLimit:   testutil.Pointer(60),
		},
	}
	require.NoError(t, c.Save(defaults))

	raw, err := os.ReadFile(c.file.path)
	require.NoError(t, err)

	var onDisk map[string]any
	require.NoError(t, yaml.Unmarshal(raw, &onDisk))
	assert.Equal(t, currentConfigSchemaVersion, onDisk["version"])
	assert.Equal(t, "game", onDisk["mode"])
	process, ok := onDisk["process"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, process["include_root"])
	game, ok := onDisk["game"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, game["confirm_mode"])
	assert.Equal(t, 2.5, game["speed"])
	assert.Equal(t, 60, game["time_limit"])
}

func TestConfigLoadAcceptsFileWithoutVersionField(t *testing.T) {
	c := newTempConfig(t)
	require.NoError(t, os.WriteFile(c.file.path, []byte("mode: lucky\n"), 0600))

	defaults, err := c.Load()

	require.NoError(t, err)
	assert.Equal(t, outbound.ModeLucky, defaults.Mode)
}

func TestConfigLoadPassesThroughUnrecognizedModeWithoutValidating(t *testing.T) {
	c := newTempConfig(t)
	require.NoError(t, os.WriteFile(c.file.path, []byte("mode: wobble\n"), 0600))

	defaults, err := c.Load()

	require.NoError(t, err, "mode validation is application/config's job, not this layer's")
	assert.Equal(t, outbound.Mode("wobble"), defaults.Mode)
}

func TestConfigLoadOnlySpeedSetLeavesOtherFieldsNil(t *testing.T) {
	c := newTempConfig(t)
	require.NoError(t, os.WriteFile(c.file.path, []byte("game:\n  speed: 3.0\n"), 0600))

	defaults, err := c.Load()

	require.NoError(t, err)
	require.NotNil(t, defaults.Game.Speed)
	assert.Equal(t, 3.0, *defaults.Game.Speed)
	assert.Nil(t, defaults.Game.ConfirmMode)
	assert.Nil(t, defaults.Game.TimeLimit)
	assert.Nil(t, defaults.Process.IncludeRoot)
}

func TestConfigLoadRejectsNewerSchemaVersion(t *testing.T) {
	c := newTempConfig(t)
	require.NoError(t, os.WriteFile(c.file.path, []byte("version: 999\n"), 0600))

	_, err := c.Load()

	require.Error(t, err)
	assert.False(t, errors.As(err, &outbound.NotFoundError{}), "a from-the-future schema version is not a missing file")
	assert.False(t, errors.As(err, &outbound.CorruptedDataError{}), "a from-the-future schema version is valid data, not corrupt")
}

func TestConfigLoadRejectsNegativeSchemaVersion(t *testing.T) {
	c := newTempConfig(t)
	require.NoError(t, os.WriteFile(c.file.path, []byte("version: -5\n"), 0600))

	_, err := c.Load()

	require.Error(t, err)
	assert.False(t, errors.As(err, &outbound.NotFoundError{}), "a negative schema version is not a missing file")
	assert.False(t, errors.As(err, &outbound.CorruptedDataError{}), "a negative schema version is a version mismatch, not corrupt YAML")
}

func TestConfigLoadAcceptsExplicitZeroSchemaVersion(t *testing.T) {
	c := newTempConfig(t)
	require.NoError(t, os.WriteFile(c.file.path, []byte("version: 0\nmode: lucky\n"), 0600))

	defaults, err := c.Load()

	require.NoError(t, err)
	assert.Equal(t, outbound.ModeLucky, defaults.Mode)
}

func TestConfigLoadRejectsFileOverMaxSize(t *testing.T) {
	c := newTempConfig(t)
	oversized := make([]byte, maxConfigFileSize+1)
	require.NoError(t, os.WriteFile(c.file.path, oversized, 0600))

	_, err := c.Load()

	var corrupted outbound.CorruptedDataError
	require.ErrorAs(t, err, &corrupted)
	assert.Contains(t, corrupted.Error(), "over the")
}

func newTempConfig(t *testing.T) *configFile {
	t.Helper()
	c := newConfigFileAt(filepath.Join(t.TempDir(), "config.yaml"))
	cf, ok := c.(*configFile)
	require.True(t, ok)
	return cf
}
