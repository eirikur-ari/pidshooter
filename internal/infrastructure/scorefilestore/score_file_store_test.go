package scorefilestore

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

func newTempStore(t *testing.T) *Store {
	t.Helper()
	return &Store{path: filepath.Join(t.TempDir(), "scores.json")}
}

func TestLoad_FileNotExist_ReturnsEmptyBoard(t *testing.T) {
	s := newTempStore(t)
	board, err := s.Load()
	require.NoError(t, err)
	assert.Equal(t, 0, board.HighScore())
}

func TestLoad_InvalidJSON_ReturnsError(t *testing.T) {
	s := newTempStore(t)
	require.NoError(t, os.WriteFile(s.path, []byte("not valid json{{{"), 0644))
	_, err := s.Load()
	assert.Error(t, err)
}

func TestSave_CreatesFile(t *testing.T) {
	s := newTempStore(t)
	require.NoError(t, s.Save(&score.Board{}))
	_, err := os.Stat(s.path)
	assert.NoError(t, err, "expected file to be created after Save")
}

func TestSave_FilePermissions(t *testing.T) {
	s := newTempStore(t)
	require.NoError(t, s.Save(&score.Board{}))
	info, err := os.Stat(s.path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestSave_Load_RoundTrip(t *testing.T) {
	s := newTempStore(t)
	board := &score.Board{}
	board.Add(score.Entry{Kills: 7, FreedMem: 4096, Speed: 2.5, Time: 60, Duration: 45.0, Date: time.Now()})

	require.NoError(t, s.Save(board))
	loaded, err := s.Load()
	require.NoError(t, err)
	assert.Equal(t, 7, loaded.HighScore())
	assert.Len(t, loaded.Scores, 1)
}

func TestSave_Load_MultipleEntries(t *testing.T) {
	s := newTempStore(t)
	board := &score.Board{}
	board.Add(score.Entry{Kills: 3, FreedMem: 1024, Speed: 2.0, Date: time.Now()})
	board.Add(score.Entry{Kills: 9, FreedMem: 8192, Speed: 3.0, Date: time.Now()})
	board.Add(score.Entry{Kills: 1, FreedMem: 512, Speed: 1.0, Date: time.Now()})

	require.NoError(t, s.Save(board))
	loaded, err := s.Load()
	require.NoError(t, err)
	assert.Equal(t, 9, loaded.HighScore())
	assert.Len(t, loaded.Scores, 3)
}

func TestSave_OverwritesPreviousFile(t *testing.T) {
	s := newTempStore(t)

	first := &score.Board{}
	first.Add(score.Entry{Kills: 2, Date: time.Now()})
	require.NoError(t, s.Save(first))

	second := &score.Board{}
	second.Add(score.Entry{Kills: 10, Date: time.Now()})
	require.NoError(t, s.Save(second))

	loaded, err := s.Load()
	require.NoError(t, err)
	assert.Equal(t, 10, loaded.HighScore())
}
