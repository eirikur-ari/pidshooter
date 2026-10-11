package filestore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// currentScoreSchemaVersion is the schema version score files are written
// with and the newest one accepted when loading.
const currentScoreSchemaVersion = 1

// maxScoreFileSize is the largest score file, in bytes, that Load accepts.
const maxScoreFileSize = 1 << 20 // 1 MiB

// scoreEncoder encodes the content of a score file.
type scoreEncoder interface {
	// encodeJSON returns the JSON encoding of value.
	encodeJSON(value any) ([]byte, error)
}

// scoreFile persists the score board to a JSON file.
type scoreFile struct {
	file    file
	encoder scoreEncoder
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
	// Version identifies scoreContent's shape. Bump currentScoreSchemaVersion
	// whenever the shape changes.
	Version int          `json:"version"`
	Scores  []scoreEntry `json:"scores"`
}

// NewScoreFile constructs an outbound.ScoreStore that persists to the
// default per-user config path. It fails if the user's home directory
// cannot be resolved.
func NewScoreFile() (outbound.ScoreStore, error) {
	f, err := newFile("highscores.json", maxScoreFileSize)
	if err != nil {
		return nil, err
	}

	return &scoreFile{file: f, encoder: fileEncoder{}}, nil
}

// newScoreFileAt constructs an outbound.ScoreStore that persists to path.
func newScoreFileAt(path string) outbound.ScoreStore {
	return &scoreFile{file: file{path: path, maxSize: maxScoreFileSize}, encoder: fileEncoder{}}
}

// Load returns the persisted score board. A missing file is reported as
// outbound.NotFoundError; a file that fails to decode is reported as
// outbound.CorruptedDataError. Any other failure, including a schema
// version mismatch, is returned unwrapped.
func (s *scoreFile) Load() (outbound.ScoreBoard, error) {
	data, err := s.file.read()
	if err != nil {
		return outbound.ScoreBoard{}, err
	}

	var c scoreContent
	if len(data) > 0 {
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&c); err != nil {
			return outbound.ScoreBoard{}, outbound.CorruptedDataError{Message: err.Error()}
		}
		if _, err := dec.Token(); !errors.Is(err, io.EOF) {
			return outbound.ScoreBoard{}, outbound.CorruptedDataError{Message: "unexpected data after the score board"}
		}
	}

	if err := schemaVersion(currentScoreSchemaVersion).validate("score file", c.Version); err != nil {
		return outbound.ScoreBoard{}, err
	}

	return toScoreBoard(c), nil
}

// Save atomically replaces the persisted score board with the given one.
func (s *scoreFile) Save(sb outbound.ScoreBoard) error {
	data, err := s.encoder.encodeJSON(toScoreContent(sb))
	if err != nil {
		return fmt.Errorf("failed to encode scores: %w", err)
	}

	return s.file.write(data)
}
