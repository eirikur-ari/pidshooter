package service

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

// --- ScoreService.loadScoreBoard ---

func TestLoadScoreBoardMapsStoredEntries(t *testing.T) {
	date := time.Now()
	svc := NewScoreService(&fake.Store{Board: outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 8, FreedMem: 1024, Speed: 2.0, Time: 30, Duration: 5, Date: date},
	}}})

	board, tracker, success := svc.loadScoreBoard()

	assert.True(t, success)
	require.Len(t, board.Scores, 1)
	assert.Equal(t, 8, tracker.HighScore)
}

func TestLoadScoreBoardErrorFallsBackToEmptyBoard(t *testing.T) {
	svc := NewScoreService(&fake.Store{LoadErr: errors.New("disk error")})

	var board *score.Board
	var tracker *score.Tracker
	var success bool
	stderr := capture.Stderr(func() {
		board, tracker, success = svc.loadScoreBoard()
	})

	assert.False(t, success)
	assert.Empty(t, board.Scores)
	assert.Equal(t, 0, tracker.HighScore)
	assert.Contains(t, stderr, "could not load scores")
}

// --- ScoreService.recordScore ---

func TestRecordScoreAddsEntryFromTracker(t *testing.T) {
	svc := NewScoreService(&fake.Store{})
	board := &score.Board{}
	tracker := &score.Tracker{Kills: 4, FreedMem: 2048}

	svc.recordScore(board, tracker, 2.5, 30, 12.5, true)

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
	svc := NewScoreService(fakeStore)
	board := &score.Board{}

	svc.recordScore(board, &score.Tracker{Kills: 1}, 0, 0, 1.0, true)

	require.NotNil(t, fakeStore.Saved)
	assert.Len(t, fakeStore.Saved.Scores, 1)
}

func TestRecordScoreSkipsSaveWhenPersistFalse(t *testing.T) {
	fakeStore := &fake.Store{}
	svc := NewScoreService(fakeStore)
	board := &score.Board{}

	svc.recordScore(board, &score.Tracker{Kills: 1}, 0, 0, 1.0, false)

	assert.Nil(t, fakeStore.Saved)
}

func TestRecordScoreSaveErrorPrintsWarning(t *testing.T) {
	fakeStore := &fake.Store{SaveErr: errors.New("disk full")}
	svc := NewScoreService(fakeStore)
	board := &score.Board{}

	stderr := capture.Stderr(func() {
		svc.recordScore(board, &score.Tracker{Kills: 1}, 0, 0, 1.0, true)
	})

	assert.Contains(t, stderr, "score not saved")
	assert.Contains(t, stderr, "disk full")
}

// --- printResults ---

func TestPrintResultsPrintsGameOverSummary(t *testing.T) {
	board := &score.Board{}

	out := capture.Output(func() {
		printResults(&score.Tracker{Kills: 3, FreedMem: 4096}, 7.5, board)
	})

	assert.Contains(t, out, "Game Over!")
	assert.Contains(t, out, "Kills: 3")
}

func TestPrintResultsDelegatesToBoardPrintScores(t *testing.T) {
	board := &score.Board{}

	out := capture.Output(func() {
		printResults(&score.Tracker{}, 0, board)
	})

	assert.Contains(t, out, "No high scores yet!")
}
