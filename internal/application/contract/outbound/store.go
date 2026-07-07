package outbound

import "github.com/eirikur-ari/pidshooter/internal/core/score"

// ScoreStore is the outbound port for persisting and retrieving the score board.
type ScoreStore interface {
	Load() (*score.Board, error)
	Save(board *score.Board) error
}
