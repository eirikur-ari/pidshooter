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

	b, tracker := toBoard(sb)

	require.Len(t, b.Scores, 1)
	assert.Equal(t, score.Entry{Kills: 5, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date}, b.Scores[0])
	assert.Equal(t, 5, tracker.HighScore)
}

func TestToBoardEmptyInput(t *testing.T) {
	b, tracker := toBoard(outbound.ScoreBoard{})

	assert.Empty(t, b.Scores)
	assert.Equal(t, 0, tracker.HighScore)
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
