// Package filescore implements outbound.ScoreStore backed by a file on disk.
package filescore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/fsutil"
)

// Store implements outbound.ScoreStore by persisting to a JSON file.
type Store struct {
	path string
}

// entry is the on-disk JSON representation of a single high score record.
type entry struct {
	Kills    int       `json:"kills"`
	FreedMem int64     `json:"freed_mem"`
	Speed    float64   `json:"speed"`
	Time     int       `json:"time_limit"`
	Duration float64   `json:"duration_secs"`
	Date     time.Time `json:"date"`
}

// board is the on-disk JSON representation of the high score table.
type board struct {
	Scores []entry `json:"scores"`
}

// NewStore constructs a *Store that persists to the default per-user
// config path. It fails if the user's home directory cannot be resolved.
func NewStore() (*Store, error) {
	path, err := defaultPath()
	if err != nil {
		return nil, err
	}
	return &Store{path: path}, nil
}

// NewStoreAt constructs a *Store that persists to the given path, without
// touching the user's default config location. Used by tests and any
// future --scores-file flag.
func NewStoreAt(path string) *Store {
	return &Store{path: path}
}

// Load reads and decodes the score board. A missing file is reported as
// outbound.NotFoundError; a file that exists but fails to decode is
// reported as outbound.CorruptedDataError. Any other read failure (e.g. a
// permission error) is returned unwrapped.
func (s *Store) Load() (outbound.ScoreBoard, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return outbound.ScoreBoard{}, outbound.NotFoundError{}
		}
		return outbound.ScoreBoard{}, err
	}

	var b board
	if err := json.Unmarshal(data, &b); err != nil {
		return outbound.ScoreBoard{}, outbound.CorruptedDataError{}
	}

	return toScoreBoard(b), nil
}

// Save encodes and persists the score board, replacing any previously
// persisted board atomically (via a temp-file-then-rename in the same
// directory) so a crash or kill mid-write can never leave a truncated or
// partial file behind.
func (s *Store) Save(sb outbound.ScoreBoard) error {
	data, err := json.MarshalIndent(toBoard(sb), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scores: %w", err)
	}

	return fsutil.WriteFileAtomic(s.path, data, 0600)
}

func toScoreBoard(b board) outbound.ScoreBoard {
	entries := make([]outbound.ScoreEntry, len(b.Scores))
	for i, e := range b.Scores {
		entries[i] = outbound.ScoreEntry{
			Kills:    e.Kills,
			FreedMem: e.FreedMem,
			Speed:    e.Speed,
			Time:     e.Time,
			Duration: e.Duration,
			Date:     e.Date,
		}
	}
	return outbound.ScoreBoard{Scores: entries}
}

func toBoard(sb outbound.ScoreBoard) board {
	entries := make([]entry, len(sb.Scores))
	for i, e := range sb.Scores {
		entries[i] = entry{
			Kills:    e.Kills,
			FreedMem: e.FreedMem,
			Speed:    e.Speed,
			Time:     e.Time,
			Duration: e.Duration,
			Date:     e.Date,
		}
	}
	return board{Scores: entries}
}

func defaultPath() (string, error) {
	dir, err := fsutil.ConfigDir("pidshooter")
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "highscores.json"), nil
}
