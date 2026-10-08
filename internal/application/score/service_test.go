package score

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

// --- Service.loadScoreBoard ---

func TestLoadScoreBoardReturnsMappedEntries(t *testing.T) {
	date := time.Now()
	svc := NewService(&testutil.FakeStore{Board: outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 8, FreedMem: 1024, Speed: 2.0, Time: 30, Duration: 5, Date: date},
	}}}, &testutil.FakeScoreReporter{})

	board, highScore, err := svc.LoadScoreBoard()

	assert.NoError(t, err)
	require.Len(t, board.Scores, 1)
	assert.Equal(t, 8, highScore)
}

func TestLoadScoreBoardReturnsNotFoundErrorWhenScoreBoardIsNotFound(t *testing.T) {
	svc := NewService(&testutil.FakeStore{LoadErr: outbound.NotFoundError{}}, &testutil.FakeScoreReporter{})

	board, highScore, err := svc.LoadScoreBoard()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	var notFound outbound.NotFoundError
	assert.ErrorAs(t, err, &notFound)
	assert.ErrorContains(t, err, "score board not loaded", "should report the underlying NotFoundError as the reason for the load failure")
	assert.Empty(t, board.Scores)
	assert.Equal(t, 0, highScore)
}

func TestLoadScoreBoardReturnsUnderlyingError(t *testing.T) {
	cause := errors.New("disk error")
	svc := NewService(&testutil.FakeStore{LoadErr: cause}, &testutil.FakeScoreReporter{})

	board, highScore, err := svc.LoadScoreBoard()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorIs(t, err, cause)
	assert.ErrorContains(t, err, "score board not loaded: disk error", "should report the underlying Error")
	assert.Empty(t, board.Scores)
	assert.Equal(t, 0, highScore)
}

// --- Service.recordScore ---

func TestRecordScoreReturnsNilWhenBoardIsSavedSuccessfully(t *testing.T) {
	fakeStore := &testutil.FakeStore{}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})
	board := score.NewBoard(nil)

	require.Nil(t, svc.RecordScore(board, score.Entry{Kills: 4, Duds: 1, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5}, nil))

	require.Len(t, board.Scores, 1)
	require.NotNil(t, fakeStore.Saved)
	assert.Len(t, fakeStore.Saved.Scores, 1)
	entry := board.Scores[0]
	assert.Equal(t, 4, entry.Kills)
	assert.Equal(t, 1, entry.Duds)
	assert.Equal(t, int64(2048), entry.FreedMem)
	assert.Equal(t, 2.5, entry.Speed)
	assert.Equal(t, 30, entry.Time)
	assert.Equal(t, 12.5, entry.Duration)
}

func TestRecordScoreDoesNotAddZeroKillEntryToBoardOrSave(t *testing.T) {
	fakeStore := &testutil.FakeStore{}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})
	board := score.NewBoard(nil)

	require.Nil(t, svc.RecordScore(board, score.Entry{Kills: 0, Duration: 1.0}, nil))

	assert.Empty(t, board.Scores, "quitting with no kills must not pollute the score board")
	require.NotNil(t, fakeStore.Saved)
	assert.Empty(t, fakeStore.Saved.Scores)
}

func TestRecordScoreMergesWithConcurrentlyPersistedEntries(t *testing.T) {
	concurrentEntry := outbound.ScoreEntry{Kills: 20, Date: time.Now()}
	fakeStore := &testutil.FakeStore{Board: outbound.ScoreBoard{Scores: []outbound.ScoreEntry{concurrentEntry}}}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})
	board := score.NewBoard(nil) // this session's board, loaded before concurrentEntry was saved by another process

	require.Nil(t, svc.RecordScore(board, score.Entry{Kills: 5, Duration: 1.0}, nil))

	require.NotNil(t, fakeStore.Saved)
	assert.Len(t, fakeStore.Saved.Scores, 2, "an entry saved by another process after this session's Load must not be discarded")
}

func TestRecordScoreSavesWhenScoreBoardWasNotFound(t *testing.T) {
	fakeStore := &testutil.FakeStore{}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})
	board := score.NewBoard(nil)

	loadErr := apperror.NewError(apperror.CodeStoreFailed, apperror.SeverityWarning, "score board not loaded", outbound.NotFoundError{})
	require.Nil(t, svc.RecordScore(board, score.Entry{Kills: 1, Duration: 1.0}, loadErr))

	require.NotNil(t, fakeStore.Saved, "a fresh (never-persisted) board should still be saved")
	assert.Len(t, fakeStore.Saved.Scores, 1)
	assert.ErrorContains(t, loadErr, "score board not loaded: not found", "should report the underlying NotFoundError as the reason for the load failure")
}

func TestRecordScoreSavesWhenScoreBoardWasCorrupted(t *testing.T) {
	fakeStore := &testutil.FakeStore{}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})
	board := score.NewBoard(nil)

	loadErr := apperror.NewError(apperror.CodeStoreFailed, apperror.SeverityWarning, "score board not loaded",
		outbound.CorruptedDataError{})
	require.Nil(t, svc.RecordScore(board, score.Entry{Kills: 1, Duration: 1.0}, loadErr))

	require.NotNil(t, fakeStore.Saved, "a corrupted (unrecoverable) board should still be saved")
	assert.Len(t, fakeStore.Saved.Scores, 1)
	assert.ErrorContains(t, loadErr, "score board not loaded: corrupted data")
}

func TestRecordScoreSkipsSaveAndReturnsWarningWhenLoadFailed(t *testing.T) {
	fakeStore := &testutil.FakeStore{}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})
	board := score.NewBoard(nil)

	cause := errors.New("disk error")
	loadErr := apperror.NewError(apperror.CodeStoreFailed, apperror.SeverityWarning, "score board not loaded", cause)

	err := svc.RecordScore(board, score.Entry{Kills: 1, Duration: 1.0}, loadErr)

	assert.Nil(t, fakeStore.Saved, "should not overwrite a file that failed to load for a real reason")
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorContains(t, err, "score board not saved")
	assert.ErrorIs(t, err, cause, "should report the original load failure as the reason the save was skipped")
}

func TestRecordScoreSaveErrorReturnsUnderlyingError(t *testing.T) {
	cause := errors.New("disk full")
	fakeStore := &testutil.FakeStore{SaveErr: cause}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})
	board := score.NewBoard(nil)

	err := svc.RecordScore(board, score.Entry{Kills: 1, Duration: 1.0}, nil)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorIs(t, err, cause)
	assert.ErrorContains(t, err, "score board not saved: disk full")
}

// --- Service.ReportResults ---

func TestReportResultsReportsSummaryViaReporter(t *testing.T) {
	board := score.NewBoard(nil)
	board.Add(score.Entry{Kills: 2, Date: time.Now()})
	reporter := &testutil.FakeScoreReporter{}
	svc := NewService(&testutil.FakeStore{}, reporter)

	svc.ReportResults(7.5, 3, 1, 4096, board)

	require.NotNil(t, reporter.Reported)
	assert.Equal(t, 3, reporter.Reported.Kills)
	assert.Equal(t, 1, reporter.Reported.Duds)
	assert.Equal(t, int64(4096), reporter.Reported.FreedMem)
	assert.Equal(t, 7.5, reporter.Reported.Duration)
	assert.True(t, reporter.Reported.IsTopScore)
	assert.Len(t, reporter.Reported.Entries, 1)
}
