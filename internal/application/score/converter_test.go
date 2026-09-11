package score

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

// --- toBoard ---

func TestToBoardMapsFields(t *testing.T) {
	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	sb := outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 5, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date},
	}}

	b := toBoard(sb)

	require.Len(t, b.Scores, 1)
	assert.Equal(t, score.Entry{Kills: 5, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date}, b.Scores[0])
	assert.Equal(t, 5, b.HighScore())
}

func TestToBoardEmptyInput(t *testing.T) {
	b := toBoard(outbound.ScoreBoard{})

	assert.Empty(t, b.Scores)
	assert.Equal(t, 0, b.HighScore())
}

// --- toScoreBoard ---

func TestToScoreBoardMapsFields(t *testing.T) {
	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	b := &score.Board{Scores: []score.Entry{
		{Kills: 5, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date},
	}}

	sb := toScoreBoard(b)

	require.Len(t, sb.Scores, 1)
	assert.Equal(t, outbound.ScoreEntry{Kills: 5, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date}, sb.Scores[0])
}

func TestToScoreBoardEmptyInput(t *testing.T) {
	sb := toScoreBoard(&score.Board{})

	assert.Empty(t, sb.Scores)
}

// --- toScoreSummary ---

func TestToScoreSummaryMapsFields(t *testing.T) {
	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	board := score.NewBoard([]score.Entry{
		{Kills: 5, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date},
	})

	summary := toScoreSummary(7.5, 3, 4096, board)

	assert.Equal(t, 3, summary.Kills)
	assert.Equal(t, int64(4096), summary.FreedMem)
	assert.Equal(t, 7.5, summary.Duration)
	require.Len(t, summary.Entries, 1)
	assert.Equal(t, outbound.ScoreEntry{Kills: 5, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date}, summary.Entries[0])
}

func TestToScoreSummaryNewHighScoreTrueWhenBeatsRecord(t *testing.T) {
	board := score.NewBoard(nil)
	board.Add(score.Entry{Kills: 3, Date: time.Now()})

	summary := toScoreSummary(1.0, 5, 0, board)

	assert.True(t, summary.NewHighScore)
}

func TestToScoreSummaryNewHighScoreFalseWhenDoesNotBeatRecord(t *testing.T) {
	board := score.NewBoard(nil)
	board.Add(score.Entry{Kills: 10, Date: time.Now()})
	board.Add(score.Entry{Kills: 3, Date: time.Now()}) // board.highScore = 10

	summary := toScoreSummary(1.0, 5, 0, board)

	assert.False(t, summary.NewHighScore)
}
