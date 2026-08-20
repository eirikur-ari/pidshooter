package score

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
	"github.com/eirikur-ari/pidshooter/internal/testutil/capture"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// --- Service.loadScoreBoard ---

func TestLoadScoreBoardMapsStoredEntries(t *testing.T) {
	date := time.Now()
	svc := NewService(&fake.Store{Board: outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 8, FreedMem: 1024, Speed: 2.0, Time: 30, Duration: 5, Date: date},
	}}})

	board, tracker, success := svc.LoadScoreBoard()

	assert.True(t, success)
	require.Len(t, board.Scores, 1)
	assert.Equal(t, 8, tracker.HighScore)
}

func TestLoadScoreBoardErrorFallsBackToEmptyBoard(t *testing.T) {
	svc := NewService(&fake.Store{LoadErr: errors.New("disk error")})

	var board *score.Board
	var tracker *score.Tracker
	var success bool
	stderr := capture.Stderr(func() {
		board, tracker, success = svc.LoadScoreBoard()
	})

	assert.False(t, success)
	assert.Empty(t, board.Scores)
	assert.Equal(t, 0, tracker.HighScore)
	assert.Contains(t, stderr, "could not load scores")
}

// --- Service.recordScore ---

func TestRecordScoreAddsEntryFromTracker(t *testing.T) {
	svc := NewService(&fake.Store{})
	board, tracker := score.NewBoard(nil)
	tracker.Kills = 4
	tracker.FreedMem = 2048

	svc.RecordScore(board, 2.5, 30, 12.5, true)

	require.Len(t, board.Scores, 1)
	entry := board.Scores[0]
	assert.Equal(t, 4, entry.Kills)
	assert.Equal(t, int64(2048), entry.FreedMem)
	assert.Equal(t, 2.5, entry.Speed)
	assert.Equal(t, 30, entry.Time)
	assert.Equal(t, 12.5, entry.Duration)
}

func TestRecordScoreSavesWhenPersistTrue(t *testing.T) {
	fakeStore := &fake.Store{}
	svc := NewService(fakeStore)
	board, tracker := score.NewBoard(nil)
	tracker.Kills = 1

	svc.RecordScore(board, 0, 0, 1.0, true)

	require.NotNil(t, fakeStore.Saved)
	assert.Len(t, fakeStore.Saved.Scores, 1)
}

func TestRecordScoreSkipsSaveWhenPersistFalse(t *testing.T) {
	fakeStore := &fake.Store{}
	svc := NewService(fakeStore)
	board, tracker := score.NewBoard(nil)
	tracker.Kills = 1

	svc.RecordScore(board, 0, 0, 1.0, false)

	assert.Nil(t, fakeStore.Saved)
}

func TestRecordScoreSaveErrorPrintsWarning(t *testing.T) {
	fakeStore := &fake.Store{SaveErr: errors.New("disk full")}
	svc := NewService(fakeStore)
	board, tracker := score.NewBoard(nil)
	tracker.Kills = 1

	stderr := capture.Stderr(func() {
		svc.RecordScore(board, 0, 0, 1.0, true)
	})

	assert.Contains(t, stderr, "score not saved")
	assert.Contains(t, stderr, "disk full")
}

// --- PrintResults ---

func TestPrintResultsPrintsGameOverSummary(t *testing.T) {
	board, tracker := score.NewBoard(nil)
	tracker.Kills = 3
	tracker.FreedMem = 4096

	out := capture.Output(func() {
		PrintResults(7.5, board)
	})

	assert.Contains(t, out, "Game Over!")
	assert.Contains(t, out, "Kills: 3")
}

func TestPrintResultsDelegatesToBoardPrintScores(t *testing.T) {
	board, _ := score.NewBoard(nil)

	out := capture.Output(func() {
		PrintResults(0, board)
	})

	assert.Contains(t, out, "No high scores yet!")
}
