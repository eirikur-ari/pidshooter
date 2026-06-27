package fake

import "github.com/eirikur-ari/pidshooter/internal/domain/score"

// Store is a test double for score.Store.
type Store struct {
	Board *score.Board
	Err   error
}

func (f *Store) Load() (*score.Board, error) {
	if f.Board == nil {
		return &score.Board{}, f.Err
	}
	return f.Board, f.Err
}

func (f *Store) Save(_ *score.Board) error { return f.Err }