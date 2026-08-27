package score

import (
	"errors"
	"fmt"
	"os"
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

// LoadScoreBoard loads the persisted score board, falling back to an empty
// board and reporting failure if the load errors for any reason other than
// no board having been persisted yet. The returned int is the board's
// current high score, seeding a caller's live session.
func (s *Service) LoadScoreBoard() (*score.Board, int, bool) {
	sb, err := s.store.Load()
	switch {
	case errors.Is(err, outbound.ErrNotFound):
		board := score.NewBoard(nil)
		return board, board.HighScore(), true
	case err != nil:
		fmt.Fprintf(os.Stderr, "warning: could not load scores: %v\n", err)
		board := score.NewBoard(nil)
		return board, board.HighScore(), false
	}
	board := toBoard(sb)
	return board, board.HighScore(), true
}

// RecordScore appends a new entry for the given session results to board, persisting the board if persist is true.
func (s *Service) RecordScore(board *score.Board, kills int, freedMem int64, speed float64, timeLimit int, duration float64, persist bool) {
	board.Add(score.Entry{
		Kills:    kills,
		FreedMem: freedMem,
		Speed:    speed,
		Time:     timeLimit,
		Duration: duration,
		Date:     time.Now(),
	})

	if persist {
		if err := s.store.Save(toScoreBoard(board)); err != nil {
			fmt.Fprintf(os.Stderr, "warning: score not saved: %v\n", err)
		}
	}
}

// PrintResults prints the end-of-game summary and the board's high scores.
func PrintResults(duration float64, kills int, freedMem int64, board *score.Board) {
	fmt.Printf("\n  Game Over! Kills: %d | Freed: %s | Time: %.1fs\n",
		kills, util.FormatBytes(freedMem), duration)
	board.PrintHighScores(kills)
}
