package filescore

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func newTempStore(t *testing.T) *Store {
	t.Helper()
	return NewStoreAt(filepath.Join(t.TempDir(), "scores.json"))
}

func TestNewStoreResolvesDefaultPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	s, err := NewStore()

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "pidshooter", "highscores.json"), s.path)
}

func TestNewStoreReturnsErrorWhenHomeUnset(t *testing.T) {
	t.Setenv("HOME", "")

	_, err := NewStore()

	assert.Error(t, err)
}

func TestLoadFileNotExistReturnsNotFoundError(t *testing.T) {
	s := newTempStore(t)
	board, err := s.Load()
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
	assert.Empty(t, board.Scores)
}

func TestLoadInvalidJSONReturnsCorruptedDataError(t *testing.T) {
	s := newTempStore(t)
	require.NoError(t, os.WriteFile(s.path, []byte("not valid json{{{"), 0644))

	_, err := s.Load()

	assert.ErrorAs(t, err, &outbound.CorruptedDataError{})
}

func TestSaveCreatesFile(t *testing.T) {
	s := newTempStore(t)
	require.NoError(t, s.Save(outbound.ScoreBoard{}))
	_, err := os.Stat(s.path)
	assert.NoError(t, err, "expected file to be created after Save")
}

func TestSaveFilePermissions(t *testing.T) {
	s := newTempStore(t)
	require.NoError(t, s.Save(outbound.ScoreBoard{}))
	info, err := os.Stat(s.path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestSaveLoadRoundTrip(t *testing.T) {
	s := newTempStore(t)
	date := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	sb := outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 7, FreedMem: 4096, Speed: 2.5, Time: 60, Duration: 45.0, Date: date},
	}}

	require.NoError(t, s.Save(sb))
	loaded, err := s.Load()
	require.NoError(t, err)
	assert.Equal(t, sb, loaded)
}

func TestSaveLoadMultipleEntries(t *testing.T) {
	s := newTempStore(t)
	date := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	sb := outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 3, FreedMem: 1024, Speed: 2.0, Time: 30, Duration: 20.0, Date: date},
		{Kills: 9, FreedMem: 8192, Speed: 3.0, Time: 45, Duration: 30.0, Date: date},
		{Kills: 1, FreedMem: 512, Speed: 1.0, Time: 15, Duration: 10.0, Date: date},
	}}

	require.NoError(t, s.Save(sb))
	loaded, err := s.Load()
	require.NoError(t, err)
	assert.Equal(t, sb, loaded)
}

func TestSaveOverwritesPreviousFile(t *testing.T) {
	s := newTempStore(t)

	first := outbound.ScoreBoard{Scores: []outbound.ScoreEntry{{Kills: 2, Date: time.Now()}}}
	require.NoError(t, s.Save(first))

	second := outbound.ScoreBoard{Scores: []outbound.ScoreEntry{{Kills: 10, Date: time.Now()}}}
	require.NoError(t, s.Save(second))

	loaded, err := s.Load()
	require.NoError(t, err)
	require.Len(t, loaded.Scores, 1)
	assert.Equal(t, 10, loaded.Scores[0].Kills)
}

func TestSaveToNestedNonexistentDirectoryCreatesParentDirs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "deeper", "scores.json")
	s := NewStoreAt(path)

	require.NoError(t, s.Save(outbound.ScoreBoard{}))

	_, err := os.Stat(path)
	assert.NoError(t, err, "Save should create the path's parent directories, not ~/.config/pidshooter")
}

func TestSaveDoesNotLeaveTempFileAfterSuccess(t *testing.T) {
	dir := t.TempDir()
	s := NewStoreAt(filepath.Join(dir, "scores.json"))

	require.NoError(t, s.Save(outbound.ScoreBoard{}))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.NotContains(t, e.Name(), ".tmp-", "no temp file should remain after a successful save")
	}
}

func TestSaveWritesJSONMatchingOnDiskSchema(t *testing.T) {
	s := newTempStore(t)
	date := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	sb := outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 7, FreedMem: 4096, Speed: 2.5, Time: 60, Duration: 45.0, Date: date},
	}}
	require.NoError(t, s.Save(sb))

	raw, err := os.ReadFile(s.path)
	require.NoError(t, err)

	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	assert.Equal(t, float64(currentSchemaVersion), onDisk["version"])
	scores, ok := onDisk["scores"].([]any)
	require.True(t, ok)
	require.Len(t, scores, 1)
	first, ok := scores[0].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, float64(7), first["kills"])
	assert.Equal(t, float64(4096), first["freed_mem"])
	assert.Equal(t, 2.5, first["speed"])
	assert.Equal(t, float64(60), first["time_limit"])
	assert.Equal(t, 45.0, first["duration_secs"])
	assert.Equal(t, date.Format(time.RFC3339Nano), first["date"])
}

func TestLoadAcceptsFileWithoutVersionField(t *testing.T) {
	s := newTempStore(t)
	require.NoError(t, os.WriteFile(s.path, []byte(`{"scores":[{"kills":5}]}`), 0600))

	board, err := s.Load()

	require.NoError(t, err)
	require.Len(t, board.Scores, 1)
	assert.Equal(t, 5, board.Scores[0].Kills)
}

func TestLoadRejectsNewerSchemaVersion(t *testing.T) {
	s := newTempStore(t)
	require.NoError(t, os.WriteFile(s.path, []byte(`{"version":999,"scores":[]}`), 0600))

	_, err := s.Load()

	require.Error(t, err)
	assert.False(t, errors.As(err, &outbound.NotFoundError{}), "a from-the-future schema version is not a missing file")
	assert.False(t, errors.As(err, &outbound.CorruptedDataError{}), "a from-the-future schema version is valid data, not corrupt")
}

func TestLoadRejectsFileOverMaxSize(t *testing.T) {
	s := newTempStore(t)
	oversized := make([]byte, maxScoreFileSize+1)
	require.NoError(t, os.WriteFile(s.path, oversized, 0600))

	_, err := s.Load()

	var corrupted outbound.CorruptedDataError
	require.ErrorAs(t, err, &corrupted)
	assert.Contains(t, corrupted.Error(), "over the")
}
