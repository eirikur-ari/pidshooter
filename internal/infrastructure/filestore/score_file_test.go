package filestore

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestNewScoreFile_ResolvesDefaultPath(t *testing.T) {
	// Given
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	expectedPath := filepath.Join(home, ".config", "pidshooter", "highscores.json")

	// When
	store, err := NewScoreFile()
	typed, ok := store.(*scoreFile)

	// Then
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, expectedPath, typed.file.path)
}

func TestNewScoreFile_ReturnsErrorWhenHomeUnset(t *testing.T) {
	// Given
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	// When
	store, err := NewScoreFile()

	// Then
	assert.Error(t, err)
	assert.Nil(t, store)
}

func TestScoreFile_Load_ReturnsNotFoundErrorWhenFileMissing(t *testing.T) {
	// Given
	store := newScoreFileFixture(t.TempDir())

	// When
	board, err := store.Load()

	// Then
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
	assert.Empty(t, board.Scores)
}

func TestScoreFile_Load_ReturnsCorruptedDataErrorForUndecodableContent(t *testing.T) {
	tests := newCorruptedScoreTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			store := newScoreFileFixture(t.TempDir())
			require.NoError(t, os.WriteFile(store.file.path, []byte(test.content), 0600))

			// When
			_, err := store.Load()

			// Then
			assert.ErrorAs(t, err, &outbound.CorruptedDataError{})
		})
	}
}

func TestScoreFile_Load_ReturnsCorruptedDataErrorWhenFileOverMaximumSize(t *testing.T) {
	// Given
	store := newScoreFileFixture(t.TempDir())
	require.NoError(t, os.WriteFile(store.file.path, make([]byte, maxScoreFileSize+1), 0600))

	// When
	_, err := store.Load()

	// Then
	assert.ErrorAs(t, err, &outbound.CorruptedDataError{})
}

func TestScoreFile_Load_ReturnsEmptyBoardForEmptyFile(t *testing.T) {
	// Given
	store := newScoreFileFixture(t.TempDir())
	require.NoError(t, os.WriteFile(store.file.path, nil, 0600))

	// When
	board, err := store.Load()

	// Then
	require.NoError(t, err)
	assert.Empty(t, board.Scores)
}

func TestScoreFile_Load_AcceptsFileWithoutVersion(t *testing.T) {
	// Given
	store := newScoreFileFixture(t.TempDir())
	require.NoError(t, os.WriteFile(store.file.path, []byte(`{"scores":[{"kills":5}]}`), 0600))

	// When
	board, err := store.Load()

	// Then
	require.NoError(t, err)
	require.Len(t, board.Scores, 1)
	assert.Equal(t, 5, board.Scores[0].Kills)
}

func TestScoreFile_Load_RejectsNewerSchemaVersionWithoutClassifyingIt(t *testing.T) {
	// Given
	store := newScoreFileFixture(t.TempDir())
	require.NoError(t, os.WriteFile(store.file.path, []byte(`{"version":999,"scores":[]}`), 0600))

	// When
	_, err := store.Load()

	// Then
	require.Error(t, err)
	assert.NotErrorAs(t, err, &outbound.NotFoundError{})
	assert.NotErrorAs(t, err, &outbound.CorruptedDataError{})
}

func TestScoreFile_Save_ReturnsErrorWhenScoreBoardCannotBeEncoded(t *testing.T) {
	// Given
	store := newScoreFileFixture(t.TempDir())
	encodeErr := errors.New("cannot encode")
	encoder := &MockFileEncoder{}
	encoder.On("encodeJSON", newScoreContentFixture()).Return(nil, encodeErr)
	store.encoder = encoder

	// When
	err := store.Save(newScoreBoardFixture())
	_, statErr := os.Stat(store.file.path)

	// Then
	assert.ErrorIs(t, err, encodeErr)
	assert.ErrorIs(t, statErr, fs.ErrNotExist)
	encoder.AssertExpectations(t)
}

func TestScoreFile_Save_ReturnsErrorWhenFileCannotBeWritten(t *testing.T) {
	// Given
	parent := filepath.Join(t.TempDir(), "parent")
	require.NoError(t, os.WriteFile(parent, nil, 0600))
	store := newScoreFileFixture(t.TempDir())
	store.file.path = filepath.Join(parent, "scores.json")

	// When
	err := store.Save(newScoreBoardFixture())

	// Then
	assert.Error(t, err)
}

func TestScoreFile_SaveLoad_RoundTripsScoreBoard(t *testing.T) {
	// Given
	store := newScoreFileFixture(t.TempDir())
	board := newScoreBoardFixture()

	// When
	saveErr := store.Save(board)
	loaded, loadErr := store.Load()

	// Then
	require.NoError(t, saveErr)
	require.NoError(t, loadErr)
	assert.Equal(t, board, loaded)
}

func TestScoreFile_Save_WritesJSONMatchingOnDiskSchema(t *testing.T) {
	// Given
	store := newScoreFileFixture(t.TempDir())

	var onDisk map[string]any
	expectedVersion := float64(currentScoreSchemaVersion)
	expectedFirstScore := map[string]any{
		"kills":         float64(7),
		"duds":          float64(2),
		"freed_mem":     float64(4096),
		"speed":         2.5,
		"time_limit":    float64(60),
		"duration_secs": 45.0,
		"date":          dateFixture().Format(time.RFC3339Nano),
	}

	// When
	saveErr := store.Save(newScoreBoardFixture())
	raw, readErr := os.ReadFile(store.file.path)
	unmarshalErr := json.Unmarshal(raw, &onDisk)
	scores, ok := onDisk["scores"].([]any)

	// Then
	require.NoError(t, saveErr)
	require.NoError(t, readErr)
	require.NoError(t, unmarshalErr)
	assert.Equal(t, expectedVersion, onDisk["version"])
	require.True(t, ok)
	require.Len(t, scores, 2)
	assert.Equal(t, expectedFirstScore, scores[0])
}

func newCorruptedScoreTestCases() []struct {
	name    string
	content string
} {
	return []struct {
		name    string
		content string
	}{
		{"invalid json", "not valid json{{{"},
		{"unknown top-level key", `{"scores":[],"bogus":1}`},
		{"unknown entry key", `{"scores":[{"kills":5,"bogus":1}]}`},
		{"data after the board", `{"scores":[]}{}`},
	}
}
