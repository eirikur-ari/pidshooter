// Package scorefilestore implements outbound.ScoreStore backed by a file on disk.
package scorefilestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

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

// Store implements outbound.ScoreStore by persisting to a JSON file.
type Store struct {
	path string
}

// NewStore returns an outbound.ScoreStore that persists to the default user config path.
func NewStore() *Store {
	return &Store{path: defaultPath()}
}

func (s *Store) Load() (outbound.ScoreBoard, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return outbound.ScoreBoard{}, outbound.ErrNotFound
		}
		return outbound.ScoreBoard{}, err
	}

	var b board
	if err := json.Unmarshal(data, &b); err != nil {
		return outbound.ScoreBoard{}, err
	}

	return toScoreBoard(b), nil
}

func (s *Store) Save(sb outbound.ScoreBoard) error {
	err := makeConfigDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(toBoard(sb), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scores: %w", err)
	}

	return os.WriteFile(s.path, data, 0600)
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

func defaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	return filepath.Join(home, ".config", "pidshooter")
}

func defaultPath() string {
	return filepath.Join(defaultDir(), "highscores.json")
}

func makeConfigDir() error {
	dir := defaultDir()
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf("could not create directory %s: %v", dir, err)
	}
	return nil
}
