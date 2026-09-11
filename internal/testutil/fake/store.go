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
	if f.LoadErr != nil {
		return outbound.ScoreBoard{}, f.LoadErr
	}
	return f.Board, nil
}

func (f *Store) Save(b outbound.ScoreBoard) error {
	scores := make([]outbound.ScoreEntry, len(b.Scores))
	copy(scores, b.Scores)
	f.Saved = &outbound.ScoreBoard{Scores: scores}
	return f.SaveErr
}
