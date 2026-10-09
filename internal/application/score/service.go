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

// BoardEntry is a recorded session.
type BoardEntry struct {
	Kills int
	// Duds is the number of targets whose process was already gone before a
	// kill could land on it.
	Duds int
	// FreedMem is the cumulative memory freed by kills, in bytes.
	FreedMem int64
	// Speed is the movement speed the session ran at.
	Speed float64
	// Time is the configured session time limit in seconds, distinct from
	// Duration.
	Time int
	// Duration is how long the session actually ran, in seconds.
	Duration float64
	Date     time.Time
}

// LoadResult holds the persisted score board.
type LoadResult struct {
	// Entries are ranked, best first.
	Entries []BoardEntry
	// HighScore is the most kills in any entry, or 0 if there are none.
	HighScore int
}

// RecordRequest holds a finished session to record on a score board.
type RecordRequest struct {
	// Entries are the score board's entries, best first.
	Entries     []BoardEntry
	Kills       int
	Duds        int
	FreedMem    int64
	LowestSpeed float64
	// TimeLimit is the session's configured time limit in seconds.
	TimeLimit int
	Duration  float64
}

// RecordResult holds the score board after a session was recorded.
type RecordResult struct {
	// Entries are the request's entries plus the recorded session, best first.
	Entries []BoardEntry
	// NewHighScore is true if the session beat every earlier entry.
	NewHighScore bool
}

// ReportRequest holds a finished session and the score board to report it against.
type ReportRequest struct {
	Duration float64
	Kills    int
	Duds     int
	FreedMem int64
	// Entries are the score board's entries, best first.
	Entries []BoardEntry
	// NewHighScore is true if the session beat every earlier entry.
	NewHighScore bool
}

// LoadScoreBoard loads the persisted score board, falling back to an
// empty board if none is persisted or the load fails. Load failures are
// classified as CodeStoreFailed and SeverityWarning.
func (s *Service) LoadScoreBoard() (LoadResult, error) {
	sb, err := s.store.Load()
	if err != nil {
		return LoadResult{}, apperror.NewError(apperror.CodeStoreFailed, apperror.SeverityWarning, "score board not loaded", err)
	}
	board := toBoard(sb)
	return LoadResult{Entries: toEntries(board.Scores), HighScore: board.HighScore()}, nil
}

// RecordScore records the session on the request's board and persists it.
// loadErr is the error from loading the board, or nil if it loaded. The board
// is persisted unless loadErr reports a failure that leaves the persisted data
// still intact and recoverable, in which case persisting is skipped and loadErr
// is returned as the reason. The result is returned even when the error is not
// nil. Any save failure is classified as CodeStoreFailed and SeverityWarning.
func (s *Service) RecordScore(request RecordRequest, loadErr error) (RecordResult, error) {
	board := score.NewBoard(toCoreEntries(request.Entries))
	entry := toCoreEntry(request)
	board.Add(entry)
	result := RecordResult{Entries: toEntries(board.Scores), NewHighScore: board.IsNewHighScore(entry.Kills)}

	err := loadErr
	if err == nil || errors.As(err, &outbound.NotFoundError{}) || errors.As(err, &outbound.CorruptedDataError{}) {
		err = s.store.Save(toScoreBoard(s.mergeWithLatest(entry, board)))
	}

	if err != nil {
		return result, apperror.NewError(apperror.CodeStoreFailed, apperror.SeverityWarning, "score board not saved", err)
	}
	return result, nil
}

// ReportResults reports the session's outcome and the board's current
// high scores via the injected outbound.ScoreReporter.
func (s *Service) ReportResults(request ReportRequest) {
	s.reporter.Report(toScoreSummary(request))
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
