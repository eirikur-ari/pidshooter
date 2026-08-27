package score

import (
	"fmt"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// Service loads and persists the score board via the injected outbound.ScoreStore.
type Service struct {
	store outbound.ScoreStore
}

// NewService constructs a Service wrapping the given outbound.ScoreStore.
func NewService(store outbound.ScoreStore) *Service {
	return &Service{store: store}
}

// LoadScoreBoard loads the persisted score board, falling back to an
// empty board if no board has been persisted yet or the load fails. The
// returned int is the board's current high score, seeding a caller's
// live session. Any error from the underlying store is returned
// unwrapped (including outbound.NotFoundError for "no board yet"); it's
// up to the caller to classify and present it.
func (s *Service) LoadScoreBoard() (*score.Board, int, error) {
	sb, err := s.store.Load()
	if err != nil {
		board := score.NewBoard(nil)
		return board, board.HighScore(), err
	}
	board := toBoard(sb)
	return board, board.HighScore(), nil
}

// RecordScore appends a new entry for the given session results to board,
// persisting the board if persist is true. Any error from the underlying
// store's Save is returned unwrapped; it's up to the caller to classify
// and present it.
func (s *Service) RecordScore(board *score.Board, kills int, freedMem int64, speed float64, timeLimit int, duration float64, persist bool) error {
	board.Add(score.Entry{
		Kills:    kills,
		FreedMem: freedMem,
		Speed:    speed,
		Time:     timeLimit,
		Duration: duration,
		Date:     time.Now(),
	})

	if persist {
		return s.store.Save(toScoreBoard(board))
	}
	return nil
}

// PrintResults prints the end-of-game summary and the board's high scores.
func PrintResults(duration float64, kills int, freedMem int64, board *score.Board) {
	fmt.Printf("\n  Game Over! Kills: %d | Freed: %s | Time: %.1fs\n",
		kills, util.FormatBytes(freedMem), duration)
	board.PrintHighScores(kills)
}
