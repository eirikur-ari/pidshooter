package filestore

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// scoreFile persists the score board to a JSON file.
type scoreFile struct {
	file file
}

// scoreEntry is the on-disk JSON representation of a single high score record.
type scoreEntry struct {
	Kills    int       `json:"kills"`
	Duds     int       `json:"duds"`
	FreedMem int64     `json:"freed_mem"`
	Speed    float64   `json:"speed"`
	Time     int       `json:"time_limit"`
	Duration float64   `json:"duration_secs"`
	Date     time.Time `json:"date"`
}

// scoreContent is the on-disk JSON representation of the high score table.
type scoreContent struct {
	// Version identifies scoreEntry's shape. Bump currentScoreSchemaVersion and
	// add a migration step in Load whenever a field is renamed or removed
	// — otherwise encoding/json silently drops or zero-fills old data on
	// the next Save.
	Version int          `json:"version"`
	Scores  []scoreEntry `json:"scores"`
}

// currentScoreSchemaVersion is the schema version this build of pidshooter
// reads and writes. See scoreContent.Version.
const currentScoreSchemaVersion = 1

// maxScoreFileSize bounds how large a score file Load will accept before
// parsing; a legitimate file holds at most 10 entries (core/score.Board's
// cap) and is a couple of KB.
const maxScoreFileSize = 1 << 20 // 1 MiB

// NewScoreFile constructs an outbound.ScoreStore that persists to the
// default per-user config path. It fails if the user's home directory
// cannot be resolved.
func NewScoreFile() (outbound.ScoreStore, error) {
	f, err := newFile("highscores.json", maxScoreFileSize)
	if err != nil {
		return nil, err
	}

	return &scoreFile{file: f}, nil
}

// newScoreFileAt constructs an outbound.ScoreStore that persists to the
// given path, without touching the user's default config location.
func newScoreFileAt(path string) outbound.ScoreStore {
	return &scoreFile{file: file{path: path, maxSize: maxScoreFileSize}}
}

// Load reads and decodes the score board. A missing file is reported as
// outbound.NotFoundError; a file that exists but fails to decode, or
// exceeds maxScoreFileSize, is reported as outbound.CorruptedDataError.
// A file written by a newer, unrecognized schema version (e.g. by a
// newer build, on a downgrade) is returned unwrapped rather than as
// outbound.CorruptedDataError. Any other read failure (e.g. a permission
// error) is also returned unwrapped.
func (s *scoreFile) Load() (outbound.ScoreBoard, error) {
	data, err := s.file.read()
	if err != nil {
		return outbound.ScoreBoard{}, err
	}

	var c scoreContent
	if err := json.Unmarshal(data, &c); err != nil {
		return outbound.ScoreBoard{}, outbound.CorruptedDataError{Message: err.Error()}
	}

	if c.Version > currentScoreSchemaVersion {
		return outbound.ScoreBoard{}, fmt.Errorf("score file schema version %d is newer than the %d this build supports", c.Version, currentScoreSchemaVersion)
	}

	return toScoreBoard(c), nil
}

// Save encodes and persists the score board, replacing any previously
// persisted board atomically, so a crash or kill mid-write can never
// leave a truncated or partial file behind.
func (s *scoreFile) Save(sb outbound.ScoreBoard) error {
	data, err := json.MarshalIndent(toScoreContent(sb), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scores: %w", err)
	}

	return s.file.write(data)
}
