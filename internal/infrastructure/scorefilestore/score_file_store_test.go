package scorefilestore

import (
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
	return &Store{path: filepath.Join(t.TempDir(), "scores.json")}
}

func TestLoadFileNotExistReturnsEmptyBoard(t *testing.T) {
	s := newTempStore(t)
	board, err := s.Load()
	require.NoError(t, err)
	assert.Empty(t, board.Scores)
}

func TestLoadInvalidJSONReturnsError(t *testing.T) {
	s := newTempStore(t)
	require.NoError(t, os.WriteFile(s.path, []byte("not valid json{{{"), 0644))
	_, err := s.Load()
	assert.Error(t, err)
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
	sb := outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 7, FreedMem: 4096, Speed: 2.5, Time: 60, Duration: 45.0, Date: time.Now()},
	}}

	require.NoError(t, s.Save(sb))
	loaded, err := s.Load()
	require.NoError(t, err)
	require.Len(t, loaded.Scores, 1)
	assert.Equal(t, 7, loaded.Scores[0].Kills)
}

func TestSaveLoadMultipleEntries(t *testing.T) {
	s := newTempStore(t)
	sb := outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 3, FreedMem: 1024, Speed: 2.0, Date: time.Now()},
		{Kills: 9, FreedMem: 8192, Speed: 3.0, Date: time.Now()},
		{Kills: 1, FreedMem: 512, Speed: 1.0, Date: time.Now()},
	}}

	require.NoError(t, s.Save(sb))
	loaded, err := s.Load()
	require.NoError(t, err)
	require.Len(t, loaded.Scores, 3)
	assert.Equal(t, 9, loaded.Scores[1].Kills)
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
