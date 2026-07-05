// Package scorefilestore implements scoredriven.Store backed by a file on disk.
package scorefilestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eirikur-ari/pidshooter/internal/domain/score"
	scoredriven "github.com/eirikur-ari/pidshooter/internal/domain/score/ports/driven"
)

// Store implements scoredriven.Store by persisting to a JSON file.
type Store struct {
	path string
}

// NewStore returns a scoredriven.Store that persists to the default user config path.
func NewStore() scoredriven.Store {
	return &Store{path: defaultPath()}
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
	err := makeConfigDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(board, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scores: %w", err)
	}

	return os.WriteFile(s.path, data, 0600)
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
