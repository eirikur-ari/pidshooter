package driven

import "github.com/eirikur-ari/pidshooter/internal/domain/score"

// Store is the driven port for persisting and retrieving the score board.
type Store interface {
	Load() (*score.Board, error)
	Save(board *score.Board) error
}
