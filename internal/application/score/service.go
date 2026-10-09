package score

import (
	"errors"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

// Service loads, records and reports scores.
type Service struct {
	store    outbound.ScoreStore
	reporter outbound.ScoreReporter
}

// NewService returns a Service that persists scores in store and reports them through reporter.
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
	return toLoadResult(sb), nil
}

// RecordScore records the session on the request's board and persists it on
// top of the latest persisted board. Nothing is persisted if the latest
// persisted board cannot be read. The result is returned even when the error
// is not nil. A failure to read or save the board is classified as
// CodeStoreFailed and SeverityWarning.
func (s *Service) RecordScore(request RecordRequest) (RecordResult, error) {
	board := score.NewBoard(toEntries(request.Entries))
	entry := toEntry(request)
	board.Add(entry)
	result := toRecordResult(board, entry.Kills)

	latest, err := s.mergeWithLatest(entry)
	if err == nil {
		err = s.store.Save(toScoreBoard(toBoardEntries(latest.Scores)))
	}

	if err != nil {
		return result, apperror.NewError(apperror.CodeStoreFailed, apperror.SeverityWarning, "score board not saved", err)
	}
	return result, nil
}

// ReportResults reports the session's outcome and the board's current high scores.
func (s *Service) ReportResults(request ReportRequest) {
	s.reporter.Report(toScoreSummary(request))
}

// mergeWithLatest returns the latest persisted board with entry added, or an error if that board cannot be read.
// A board that is missing or corrupted counts as empty.
func (s *Service) mergeWithLatest(entry score.Entry) (*score.Board, error) {
	sb, err := s.store.Load()
	if err != nil && !errors.As(err, &outbound.NotFoundError{}) && !errors.As(err, &outbound.CorruptedDataError{}) {
		return nil, err
	}
	latest := toBoard(sb)
	latest.Add(entry)
	return latest, nil
}
