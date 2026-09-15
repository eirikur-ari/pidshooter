package score

import (
	"errors"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

// Service loads and persists the score board via the injected
// outbound.ScoreStore, and reports session results via the injected
// outbound.ScoreReporter.
type Service struct {
	store    outbound.ScoreStore
	reporter outbound.ScoreReporter
}

// NewService constructs a Service wrapping the given outbound.ScoreStore and outbound.ScoreReporter.
func NewService(store outbound.ScoreStore, reporter outbound.ScoreReporter) *Service {
	return &Service{store: store, reporter: reporter}
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
//
// Before saving, the new entry is applied to a freshly reloaded copy of
// the persisted board rather than to board as loaded at session start, so
// a concurrent save from another pidshooter process in the meantime isn't
// silently discarded by this one overwriting the whole file.
func (s *Service) RecordScore(board *score.Board, kills, duds int, freedMem int64, speed float64, timeLimit int, duration float64, err error) error {
	entry := score.Entry{
		Kills:    kills,
		Duds:     duds,
		FreedMem: freedMem,
		Speed:    speed,
		Time:     timeLimit,
		Duration: duration,
		Date:     time.Now(),
	}
	board.Add(entry)

	if err == nil || errors.As(err, &outbound.NotFoundError{}) || errors.As(err, &outbound.CorruptedDataError{}) {
		err = s.store.Save(toScoreBoard(s.mergeWithLatest(entry, board)))
	}

	if err != nil {
		return apperror.NewError(apperror.CodeScoreSaveFailed, apperror.SeverityWarning, "score board not saved", err)
	}
	return nil
}

// ReportResults reports the session's outcome and the board's current
// high scores via the injected outbound.ScoreReporter.
func (s *Service) ReportResults(duration float64, kills, duds int, freedMem int64, board *score.Board) {
	s.reporter.Report(toScoreSummary(duration, kills, duds, freedMem, board))
}

// mergeWithLatest re-loads the currently persisted board and applies entry
// to that fresh copy, so a save from another process that landed after
// this session started isn't overwritten. Falls back to board (already
// updated with entry) if the reload itself fails for a reason other than
// "nothing persisted yet" or "persisted data was unparseable" — both
// treated the same as a fresh board, matching RecordScore's own
// overwrite-safety gate above.
func (s *Service) mergeWithLatest(entry score.Entry, board *score.Board) *score.Board {
	sb, err := s.store.Load()
	if err != nil && !errors.As(err, &outbound.NotFoundError{}) && !errors.As(err, &outbound.CorruptedDataError{}) {
		return board
	}
	latest := toBoard(sb)
	latest.Add(entry)
	return latest
}
