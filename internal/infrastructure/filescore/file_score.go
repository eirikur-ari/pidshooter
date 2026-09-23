package filescore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/fsutil"
)

// fileScore persists the score board to a JSON file.
type fileScore struct {
	path string
}

// fileEntry is the on-disk JSON representation of a single high score record.
type fileEntry struct {
	Kills    int       `json:"kills"`
	Duds     int       `json:"duds"`
	FreedMem int64     `json:"freed_mem"`
	Speed    float64   `json:"speed"`
	Time     int       `json:"time_limit"`
	Duration float64   `json:"duration_secs"`
	Date     time.Time `json:"date"`
}

// fileContent is the on-disk JSON representation of the high score table.
type fileContent struct {
	// Version identifies fileEntry's shape. Bump currentSchemaVersion and
	// add a migration step in Load whenever a field is renamed or removed
	// — otherwise encoding/json silently drops or zero-fills old data on
	// the next Save.
	Version int         `json:"version"`
	Scores  []fileEntry `json:"scores"`
}

// currentSchemaVersion is the schema version this build of pidshooter
// reads and writes. See fileContent.Version.
const currentSchemaVersion = 1

// maxFileSize bounds how large a score file Load will accept before
// parsing; a legitimate file holds at most 10 entries (core/score.Board's
// cap) and is a couple of KB.
const maxFileSize = 1 << 20 // 1 MiB

// NewFileScore constructs an outbound.ScoreStore that persists to the
// default per-user config path. It fails if the user's home directory
// cannot be resolved.
func NewFileScore() (outbound.ScoreStore, error) {
	path, err := defaultPath()
	if err != nil {
		return nil, err
	}
	return &fileScore{path: path}, nil
}

// newFileScoreAt constructs an outbound.ScoreStore that persists to the
// given path, without touching the user's default config location.
func newFileScoreAt(path string) outbound.ScoreStore {
	return &fileScore{path: path}
}

// Load reads and decodes the score board. A missing file is reported as
// outbound.NotFoundError; a file that exists but fails to decode, or
// exceeds maxFileSize, is reported as outbound.CorruptedDataError.
// A file written by a newer, unrecognized schema version (e.g. by a
// newer build, on a downgrade) is returned unwrapped rather than as
// outbound.CorruptedDataError. Any other read failure (e.g. a permission
// error) is also returned unwrapped.
func (f *fileScore) Load() (outbound.ScoreBoard, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return outbound.ScoreBoard{}, outbound.NotFoundError{}
		}
		return outbound.ScoreBoard{}, err
	}
	if len(data) > maxFileSize {
		msg := fmt.Sprintf("score file is %d bytes, over the %d byte limit", len(data), maxFileSize)
		return outbound.ScoreBoard{}, outbound.CorruptedDataError{Message: msg}
	}

	var content fileContent
	if err := json.Unmarshal(data, &content); err != nil {
		return outbound.ScoreBoard{}, outbound.CorruptedDataError{Message: err.Error()}
	}
	if content.Version > currentSchemaVersion {
		return outbound.ScoreBoard{}, fmt.Errorf("score file schema version %d is newer than the %d this build supports", content.Version, currentSchemaVersion)
	}

	return toScoreBoard(content), nil
}

// Save encodes and persists the score board, replacing any previously
// persisted board atomically, so a crash or kill mid-write can never
// leave a truncated or partial file behind.
func (f *fileScore) Save(sb outbound.ScoreBoard) error {
	data, err := json.MarshalIndent(toFileContent(sb), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scores: %w", err)
	}

	return fsutil.WriteFileAtomic(f.path, data, 0600)
}

func toScoreBoard(fc fileContent) outbound.ScoreBoard {
	entries := make([]outbound.ScoreEntry, len(fc.Scores))
	for i, e := range fc.Scores {
		entries[i] = outbound.ScoreEntry{
			Kills:    e.Kills,
			Duds:     e.Duds,
			FreedMem: e.FreedMem,
			Speed:    e.Speed,
			Time:     e.Time,
			Duration: e.Duration,
			Date:     e.Date,
		}
	}
	return outbound.ScoreBoard{Scores: entries}
}

func toFileContent(sb outbound.ScoreBoard) fileContent {
	entries := make([]fileEntry, len(sb.Scores))
	for i, e := range sb.Scores {
		entries[i] = fileEntry{
			Kills:    e.Kills,
			Duds:     e.Duds,
			FreedMem: e.FreedMem,
			Speed:    e.Speed,
			Time:     e.Time,
			Duration: e.Duration,
			Date:     e.Date,
		}
	}
	return fileContent{Version: currentSchemaVersion, Scores: entries}
}

func defaultPath() (string, error) {
	dir, err := fsutil.ConfigDir("pidshooter")
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "highscores.json"), nil
}
