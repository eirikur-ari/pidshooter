package fake

import "github.com/eirikur-ari/pidshooter/internal/core/score"

// Store is a test double for spi.Store.
type Store struct {
	Board   *score.Board
	LoadErr error
	SaveErr error
	Saved   *score.Board // captured by the most recent Save call
}

func (f *Store) Load() (*score.Board, error) {
	if f.Board == nil {
		return &score.Board{}, f.LoadErr
	}
	return f.Board, f.LoadErr
}

func (f *Store) Save(b *score.Board) error {
	f.Saved = b
	return f.SaveErr
}
