package score

import (
	"errors"

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
// empty board if none is persisted or the load fails. The returned int is
// the board's high score. Load failures are classified as
// CodeScoreLoadFailed and SeverityWarning.
func (s *Service) LoadScoreBoard() (*score.Board, int, error) {
	sb, err := s.store.Load()
	if err != nil {
		board := score.NewBoard(nil)
		return board, board.HighScore(), apperror.NewError(apperror.CodeScoreLoadFailed, apperror.SeverityWarning, "score board not loaded", err)
	}
	board := toBoard(sb)
	return board, board.HighScore(), nil
}

// RecordScore appends entry to board. board is persisted unless err (the
// error LoadScoreBoard returned for this session) reports a load failure
// that leaves the persisted data still intact and recoverable, in which
// case persisting is skipped and err is returned as the reason. Any save
// failure is classified as CodeScoreSaveFailed and SeverityWarning.
func (s *Service) RecordScore(board *score.Board, entry score.Entry, err error) error {
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

// mergeWithLatest re-loads the persisted board and appends entry to it,
// falling back to board if the reload fails for any reason other than no
// board being persisted yet or the persisted data being unparseable.
func (s *Service) mergeWithLatest(entry score.Entry, board *score.Board) *score.Board {
	sb, err := s.store.Load()
	if err != nil && !errors.As(err, &outbound.NotFoundError{}) && !errors.As(err, &outbound.CorruptedDataError{}) {
		return board
	}
	latest := toBoard(sb)
	latest.Add(entry)
	return latest
}
