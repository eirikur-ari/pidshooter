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
	"github.com/eirikur-ari/pidshooter/internal/testutil/capture"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// --- Service.loadScoreBoard ---

func TestLoadScoreBoardReturnsMappedEntries(t *testing.T) {
	date := time.Now()
	svc := NewService(&fake.Store{Board: outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 8, FreedMem: 1024, Speed: 2.0, Time: 30, Duration: 5, Date: date},
	}}})

	board, highScore, err := svc.LoadScoreBoard()

	assert.NoError(t, err)
	require.Len(t, board.Scores, 1)
	assert.Equal(t, 8, highScore)
}

func TestLoadScoreBoardReturnsNotFoundErrorWhenScoreBoardIsNotFound(t *testing.T) {
	svc := NewService(&fake.Store{LoadErr: outbound.NotFoundError{}})

	board, highScore, err := svc.LoadScoreBoard()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeScoreLoadFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	var notFound outbound.NotFoundError
	assert.ErrorAs(t, err, &notFound)
	assert.ErrorContains(t, err, "score board not loaded", "should report the underlying NotFoundError as the reason for the load failure")
	assert.Empty(t, board.Scores)
	assert.Equal(t, 0, highScore)
}

func TestLoadScoreBoardReturnsUnderlyingError(t *testing.T) {
	cause := errors.New("disk error")
	svc := NewService(&fake.Store{LoadErr: cause})

	board, highScore, err := svc.LoadScoreBoard()

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeScoreLoadFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorIs(t, err, cause)
	assert.ErrorContains(t, err, "score board not loaded: disk error", "should report the underlying Error")
	assert.Empty(t, board.Scores)
	assert.Equal(t, 0, highScore)
}

// --- Service.recordScore ---

func TestRecordScoreReturnsNilWhenBoardIsSavedSuccessfully(t *testing.T) {
	fakeStore := &fake.Store{}
	svc := NewService(fakeStore)
	board := score.NewBoard(nil)

	require.Nil(t, svc.RecordScore(board, 4, 2048, 2.5, 30, 12.5, nil))

	require.Len(t, board.Scores, 1)
	require.NotNil(t, fakeStore.Saved)
	assert.Len(t, fakeStore.Saved.Scores, 1)
	entry := board.Scores[0]
	assert.Equal(t, 4, entry.Kills)
	assert.Equal(t, int64(2048), entry.FreedMem)
	assert.Equal(t, 2.5, entry.Speed)
	assert.Equal(t, 30, entry.Time)
	assert.Equal(t, 12.5, entry.Duration)
}

func TestRecordScoreSavesWhenScoreBoardWasNotFound(t *testing.T) {
	fakeStore := &fake.Store{}
	svc := NewService(fakeStore)
	board := score.NewBoard(nil)

	loadErr := apperror.NewError(apperror.CodeScoreLoadFailed, apperror.SeverityWarning, "score board not loaded", outbound.NotFoundError{})
	require.Nil(t, svc.RecordScore(board, 1, 0, 0, 0, 1.0, loadErr))

	require.NotNil(t, fakeStore.Saved, "a fresh (never-persisted) board should still be saved")
	assert.Len(t, fakeStore.Saved.Scores, 1)
	assert.ErrorContains(t, loadErr, "score board not loaded: not found", "should report the underlying NotFoundError as the reason for the load failure")
}

func TestRecordScoreSavesWhenScoreBoardWasCorrupted(t *testing.T) {
	fakeStore := &fake.Store{}
	svc := NewService(fakeStore)
	board := score.NewBoard(nil)

	loadErr := apperror.NewError(apperror.CodeScoreLoadFailed, apperror.SeverityWarning, "score board not loaded",
		outbound.CorruptedDataError{})
	require.Nil(t, svc.RecordScore(board, 1, 0, 0, 0, 1.0, loadErr))

	require.NotNil(t, fakeStore.Saved, "a corrupted (unrecoverable) board should still be saved")
	assert.Len(t, fakeStore.Saved.Scores, 1)
	assert.ErrorContains(t, loadErr, "score board not loaded: corrupted data")
}

func TestRecordScoreSkipsSaveAndReturnsWarningWhenLoadFailed(t *testing.T) {
	fakeStore := &fake.Store{}
	svc := NewService(fakeStore)
	board := score.NewBoard(nil)

	cause := errors.New("disk error")
	loadErr := apperror.NewError(apperror.CodeScoreLoadFailed, apperror.SeverityWarning, "score board not loaded", cause)

	err := svc.RecordScore(board, 1, 0, 0, 0, 1.0, loadErr)

	assert.Nil(t, fakeStore.Saved, "should not overwrite a file that failed to load for a real reason")
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeScoreSaveFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorIs(t, err, cause, "should report the original load failure as the reason the save was skipped")
}

func TestRecordScoreSaveErrorReturnsUnderlyingError(t *testing.T) {
	cause := errors.New("disk full")
	fakeStore := &fake.Store{SaveErr: cause}
	svc := NewService(fakeStore)
	board := score.NewBoard(nil)

	err := svc.RecordScore(board, 1, 0, 0, 0, 1.0, nil)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeScoreSaveFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorIs(t, err, cause)
}

// --- PrintResults ---

func TestPrintResultsPrintsGameOverSummary(t *testing.T) {
	board := score.NewBoard(nil)

	out := capture.Output(func() {
		PrintResults(7.5, 3, 4096, board)
	})

	assert.Contains(t, out, "Game Over!")
	assert.Contains(t, out, "Kills: 3")
}

func TestPrintResultsDelegatesToBoardPrintScores(t *testing.T) {
	board := score.NewBoard(nil)

	out := capture.Output(func() {
		PrintResults(0, 0, 0, board)
	})

	assert.Contains(t, out, "No high scores yet!")
}
