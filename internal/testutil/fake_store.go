package testutil

import "github.com/eirikur-ari/pidshooter/internal/domain/score"

// FakeStore is a test double for score.Store.
type FakeStore struct {
	Board *score.Board
	Err   error
}

func (f *FakeStore) Load() (*score.Board, error) {
	if f.Board == nil {
		return &score.Board{}, f.Err
	}
	return f.Board, f.Err
}

func (f *FakeStore) Save(_ *score.Board) error { return f.Err }
