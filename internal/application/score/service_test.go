package score

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

// --- Service.loadScoreBoard ---

func TestLoadScoreBoardReturnsMappedEntries(t *testing.T) {
	date := time.Now()
	svc := NewService(&testutil.FakeStore{Board: outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 8, FreedMem: 1024, Speed: 2.0, Time: 30, Duration: 5, Date: date},
	}}}, &testutil.FakeScoreReporter{})

	result, err := svc.LoadScoreBoard()

	assert.NoError(t, err)
	assert.Equal(t, []BoardEntry{{Kills: 8, FreedMem: 1024, Speed: 2.0, Time: 30, Duration: 5, Date: date}}, result.Entries)
	assert.Equal(t, 8, result.HighScore)
}

func TestLoadScoreBoardReturnsNotFoundErrorWhenScoreBoardIsNotFound(t *testing.T) {
	svc := NewService(&testutil.FakeStore{LoadErr: outbound.NotFoundError{}}, &testutil.FakeScoreReporter{})

	result, err := svc.LoadScoreBoard()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	var notFound outbound.NotFoundError
	assert.ErrorAs(t, err, &notFound)
	assert.ErrorContains(t, err, "score board not loaded", "should report the underlying NotFoundError as the reason for the load failure")
	assert.Empty(t, result.Entries)
	assert.Equal(t, 0, result.HighScore)
}

func TestLoadScoreBoardReturnsUnderlyingError(t *testing.T) {
	cause := errors.New("disk error")
	svc := NewService(&testutil.FakeStore{LoadErr: cause}, &testutil.FakeScoreReporter{})

	result, err := svc.LoadScoreBoard()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorIs(t, err, cause)
	assert.ErrorContains(t, err, "score board not loaded: disk error", "should report the underlying Error")
	assert.Empty(t, result.Entries)
	assert.Equal(t, 0, result.HighScore)
}

// --- Service.recordScore ---

func TestRecordScoreReturnsNilWhenBoardIsSavedSuccessfully(t *testing.T) {
	fakeStore := &testutil.FakeStore{}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})
	request := RecordRequest{Kills: 4, Duds: 1, FreedMem: 2048, LowestSpeed: 2.5, TimeLimit: 30, Duration: 12.5}

	result, err := svc.RecordScore(request, nil)

	require.NoError(t, err)
	require.Len(t, result.Entries, 1)
	require.NotNil(t, fakeStore.Saved)
	assert.Len(t, fakeStore.Saved.Scores, 1)
	entry := result.Entries[0]
	assert.Equal(t, 4, entry.Kills)
	assert.Equal(t, 1, entry.Duds)
	assert.Equal(t, int64(2048), entry.FreedMem)
	assert.Equal(t, 2.5, entry.Speed)
	assert.Equal(t, 30, entry.Time)
	assert.Equal(t, 12.5, entry.Duration)
	assert.False(t, entry.Date.IsZero())
}

func TestRecordScoreDoesNotAddZeroKillEntryToBoardOrSave(t *testing.T) {
	fakeStore := &testutil.FakeStore{}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})

	result, err := svc.RecordScore(RecordRequest{Kills: 0, Duration: 1.0}, nil)

	require.NoError(t, err)
	assert.Empty(t, result.Entries, "quitting with no kills must not pollute the score board")
	require.NotNil(t, fakeStore.Saved)
	assert.Empty(t, fakeStore.Saved.Scores)
}

func TestRecordScoreMergesWithConcurrentlyPersistedEntries(t *testing.T) {
	concurrentEntry := outbound.ScoreEntry{Kills: 20, Date: time.Now()}
	fakeStore := &testutil.FakeStore{Board: outbound.ScoreBoard{Scores: []outbound.ScoreEntry{concurrentEntry}}}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})

	// the request's board was loaded before concurrentEntry was saved by another process
	result, err := svc.RecordScore(RecordRequest{Kills: 5, Duration: 1.0}, nil)

	require.NoError(t, err)
	require.NotNil(t, fakeStore.Saved)
	assert.Len(t, fakeStore.Saved.Scores, 2, "an entry saved by another process after this session's Load must not be discarded")
	assert.Len(t, result.Entries, 1, "the result holds this session's board, not the merged one that was saved")
}

func TestRecordScoreSavesSessionBoardWhenReloadingPersistedBoardFails(t *testing.T) {
	fakeStore := &testutil.FakeStore{LoadErr: errors.New("disk error")}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})

	_, err := svc.RecordScore(RecordRequest{Entries: []BoardEntry{{Kills: 3}}, Kills: 5, Duration: 1.0}, nil)

	require.NoError(t, err)
	require.NotNil(t, fakeStore.Saved)
	assert.Len(t, fakeStore.Saved.Scores, 2, "the session's own board should be saved when the persisted one cannot be reloaded")
}

func TestRecordScoreSavesWhenScoreBoardWasNotFound(t *testing.T) {
	fakeStore := &testutil.FakeStore{}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})

	loadErr := apperror.NewError(apperror.CodeStoreFailed, apperror.SeverityWarning, "score board not loaded", outbound.NotFoundError{})
	_, err := svc.RecordScore(RecordRequest{Kills: 1, Duration: 1.0}, loadErr)

	require.NoError(t, err)
	require.NotNil(t, fakeStore.Saved, "a fresh (never-persisted) board should still be saved")
	assert.Len(t, fakeStore.Saved.Scores, 1)
	assert.ErrorContains(t, loadErr, "score board not loaded: not found", "should report the underlying NotFoundError as the reason for the load failure")
}

func TestRecordScoreSavesWhenScoreBoardWasCorrupted(t *testing.T) {
	fakeStore := &testutil.FakeStore{}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})

	loadErr := apperror.NewError(apperror.CodeStoreFailed, apperror.SeverityWarning, "score board not loaded",
		outbound.CorruptedDataError{})
	_, err := svc.RecordScore(RecordRequest{Kills: 1, Duration: 1.0}, loadErr)

	require.NoError(t, err)
	require.NotNil(t, fakeStore.Saved, "a corrupted (unrecoverable) board should still be saved")
	assert.Len(t, fakeStore.Saved.Scores, 1)
	assert.ErrorContains(t, loadErr, "score board not loaded: corrupted data")
}

func TestRecordScoreSkipsSaveAndReturnsWarningWhenLoadFailed(t *testing.T) {
	fakeStore := &testutil.FakeStore{}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})

	cause := errors.New("disk error")
	loadErr := apperror.NewError(apperror.CodeStoreFailed, apperror.SeverityWarning, "score board not loaded", cause)

	result, err := svc.RecordScore(RecordRequest{Kills: 1, Duration: 1.0}, loadErr)

	assert.Nil(t, fakeStore.Saved, "should not overwrite a file that failed to load for a real reason")
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorContains(t, err, "score board not saved")
	assert.ErrorIs(t, err, cause, "should report the original load failure as the reason the save was skipped")
	assert.Len(t, result.Entries, 1, "the result is returned even though the save was skipped")
}

func TestRecordScoreSaveErrorReturnsUnderlyingError(t *testing.T) {
	cause := errors.New("disk full")
	fakeStore := &testutil.FakeStore{SaveErr: cause}
	svc := NewService(fakeStore, &testutil.FakeScoreReporter{})

	result, err := svc.RecordScore(RecordRequest{Kills: 1, Duration: 1.0}, nil)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorIs(t, err, cause)
	assert.ErrorContains(t, err, "score board not saved: disk full")
	assert.Len(t, result.Entries, 1, "the result is returned even though the save failed")
}

func TestRecordScoreReportsNewHighScoreOnlyWhenKillsBeatEveryEntry(t *testing.T) {
	tests := []struct {
		name     string
		kills    int
		expected bool
	}{
		{"more kills than the best entry", 5, true},
		{"as many kills as the best entry", 3, false},
		{"fewer kills than the best entry", 2, false},
		{"no kills", 0, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			svc := NewService(&testutil.FakeStore{}, &testutil.FakeScoreReporter{})
			request := RecordRequest{Entries: []BoardEntry{{Kills: 3}}, Kills: test.kills}

			// When
			result, err := svc.RecordScore(request, nil)

			// Then
			require.NoError(t, err)
			assert.Equal(t, test.expected, result.NewHighScore)
		})
	}
}

// --- Service.ReportResults ---

func TestReportResultsReportsSummaryViaReporter(t *testing.T) {
	reporter := &testutil.FakeScoreReporter{}
	svc := NewService(&testutil.FakeStore{}, reporter)

	svc.ReportResults(ReportRequest{Duration: 7.5, Kills: 3, Duds: 1, FreedMem: 4096, Entries: []BoardEntry{{Kills: 2}}, NewHighScore: true})

	require.NotNil(t, reporter.Reported)
	assert.Equal(t, 3, reporter.Reported.Kills)
	assert.Equal(t, 1, reporter.Reported.Duds)
	assert.Equal(t, int64(4096), reporter.Reported.FreedMem)
	assert.Equal(t, 7.5, reporter.Reported.Duration)
	assert.True(t, reporter.Reported.IsTopScore)
	assert.Len(t, reporter.Reported.Entries, 1)
}
