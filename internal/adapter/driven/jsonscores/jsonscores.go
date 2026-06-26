// Package jsonscores implements score.Store backed by a JSON file on disk.
package jsonscores

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eirikur-ari/pidshooter/internal/domain/score"
)

// Store implements score.Store by persisting to a JSON file.
type Store struct {
	path string
}

// NewStore returns a score.Store that persists to the default user config path.
func NewStore() score.Store {
	return &Store{path: defaultPath()}
}

func defaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".config", "pidshooter")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "highscores.json")
}

func (s *Store) Load() (*score.Board, error) {
	b := &score.Board{}
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return b, nil
		}
		return b, err
	}
	if err := json.Unmarshal(data, b); err != nil {
		return b, err
	}
	return b, nil
}

func (s *Store) Save(board *score.Board) error {
	data, err := json.MarshalIndent(board, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scores: %w", err)
	}
	return os.WriteFile(s.path, data, 0644)
}
