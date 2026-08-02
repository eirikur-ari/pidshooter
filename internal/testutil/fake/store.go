package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// Store is a test double for outbound.ScoreStore.
type Store struct {
	Board   outbound.ScoreBoard
	LoadErr error
	SaveErr error
	Saved   *outbound.ScoreBoard // captured by the most recent Save call, nil if Save was never called
}

func (f *Store) Load() (outbound.ScoreBoard, error) {
	return f.Board, f.LoadErr
}

func (f *Store) Save(b outbound.ScoreBoard) error {
	f.Saved = &b
	return f.SaveErr
}
