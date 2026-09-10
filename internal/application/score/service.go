package score

import (
	"errors"
	"fmt"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
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
// live session. Any load failure — including outbound.NotFoundError for
// "no board yet" — is classified as CodeScoreLoadFailed and SeverityWarning;
// pass the returned error straight to RecordScore so it can decide whether
// persisting afterward is safe.
func (s *Service) LoadScoreBoard() (*score.Board, int, error) {
	sb, err := s.store.Load()
	if err != nil {
		board := score.NewBoard(nil)
		return board, board.HighScore(), apperror.NewError(apperror.CodeScoreLoadFailed, apperror.SeverityWarning, "score board not loaded", err)
	}
	board := toBoard(sb)
	return board, board.HighScore(), nil
}

// RecordScore appends a new entry for the given session results to board.
// err is whatever LoadScoreBoard returned for this session: the board is
// persisted when err is nil, wraps outbound.NotFoundError (a fresh
// install), or wraps outbound.CorruptedDataError (the persisted data
// could not be parsed, so there's nothing left to protect by refusing to
// overwrite it) — safe to write in all three cases — and the save is
// skipped, reporting err as the reason, for any other load failure (e.g.
// a permission or I/O error), since the data may still be intact and
// recoverable. A save failure is classified as CodeScoreSaveFailed and
// SeverityWarning either way.
func (s *Service) RecordScore(board *score.Board, kills int, freedMem int64, speed float64, timeLimit int, duration float64, err error) error {
	board.Add(score.Entry{
		Kills:    kills,
		FreedMem: freedMem,
		Speed:    speed,
		Time:     timeLimit,
		Duration: duration,
		Date:     time.Now(),
	})

	if err == nil || errors.As(err, &outbound.NotFoundError{}) || errors.As(err, &outbound.CorruptedDataError{}) {
		err = s.store.Save(toScoreBoard(board))
	}

	if err != nil {
		return apperror.NewError(apperror.CodeScoreSaveFailed, apperror.SeverityWarning, "score board not saved", err)
	}
	return nil
}

// PrintResults prints the end-of-game summary and the board's high scores.
func PrintResults(duration float64, kills int, freedMem int64, board *score.Board) {
	fmt.Printf("\n  Game Over! Kills: %d | Freed: %s | Time: %.1fs\n",
		kills, util.FormatBytes(freedMem), duration)
	board.PrintHighScores(kills)
}
