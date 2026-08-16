package service

import (
	"fmt"
	"os"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// ScoreService loads and persists the score board via the injected outbound.ScoreStore.
type ScoreService struct {
	store outbound.ScoreStore
}

// NewScoreService constructs a ScoreService wrapping the given outbound.ScoreStore.
func NewScoreService(store outbound.ScoreStore) *ScoreService {
	return &ScoreService{store: store}
}

func (s *ScoreService) loadScoreBoard() (*score.Board, *score.Tracker, bool) {
	sb, err := s.store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load scores: %v\n", err)
		board, tracker := score.NewBoard(nil)
		return board, tracker, false
	}
	board, tracker := toBoard(sb)
	return board, tracker, true
}

func (s *ScoreService) recordScore(board *score.Board, tracker *score.Tracker, speed float64, timeLimit int, duration float64, persist bool) {
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

func printResults(tracker *score.Tracker, duration float64, board *score.Board) {
	fmt.Printf("\n  Game Over! Kills: %d | Freed: %s | Time: %.1fs\n",
		tracker.Kills, util.FormatBytes(tracker.FreedMem), duration)
	board.PrintScores(tracker)
}
