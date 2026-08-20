package score

import (
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
// board and reporting failure if the load errors.
func (s *Service) LoadScoreBoard() (*score.Board, *score.Tracker, bool) {
	sb, err := s.store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load scores: %v\n", err)
		board, tracker := score.NewBoard(nil)
		return board, tracker, false
	}
	board, tracker := toBoard(sb)
	return board, tracker, true
}

// RecordScore appends the tracker's results to board as a new entry, persisting the board if persist is true.
func (s *Service) RecordScore(board *score.Board, speed float64, timeLimit int, duration float64, persist bool) {
	tracker := board.Tracker()
	board.Add(score.Entry{
		Kills:    tracker.Kills,
		FreedMem: tracker.FreedMem,
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
func PrintResults(duration float64, board *score.Board) {
	tracker := board.Tracker()
	fmt.Printf("\n  Game Over! Kills: %d | Freed: %s | Time: %.1fs\n",
		tracker.Kills, util.FormatBytes(tracker.FreedMem), duration)
	board.PrintHighScores()
}
