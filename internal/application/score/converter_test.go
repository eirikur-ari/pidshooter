package score

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

// --- toCoreEntry ---

func TestToCoreEntryMapsFields(t *testing.T) {
	request := RecordRequest{Kills: 5, Duds: 2, FreedMem: 2048, LowestSpeed: 2.5, TimeLimit: 30, Duration: 12.5}

	entry := toCoreEntry(request)

	assert.Equal(t, 5, entry.Kills)
	assert.Equal(t, 2, entry.Duds)
	assert.Equal(t, int64(2048), entry.FreedMem)
	assert.Equal(t, 2.5, entry.Speed)
	assert.Equal(t, 30, entry.Time)
	assert.Equal(t, 12.5, entry.Duration)
	assert.False(t, entry.Date.IsZero())
}

// --- toCoreEntries ---

func TestToCoreEntriesMapsFieldsInOrder(t *testing.T) {
	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	entries := toCoreEntries([]BoardEntry{
		{Kills: 5, Duds: 1, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date},
		{Kills: 3},
	})

	assert.Equal(t, []score.Entry{
		{Kills: 5, Duds: 1, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date},
		{Kills: 3},
	}, entries)
}

func TestToCoreEntriesEmptyInput(t *testing.T) {
	assert.Empty(t, toCoreEntries(nil))
}

// --- toEntries ---

func TestToEntriesMapsFieldsInOrder(t *testing.T) {
	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	entries := toEntries([]score.Entry{
		{Kills: 5, Duds: 1, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date},
		{Kills: 3},
	})

	assert.Equal(t, []BoardEntry{
		{Kills: 5, Duds: 1, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date},
		{Kills: 3},
	}, entries)
}

func TestToEntriesEmptyInput(t *testing.T) {
	assert.Empty(t, toEntries(nil))
}

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
	request := ReportRequest{
		Duration: 7.5, Kills: 3, Duds: 2, FreedMem: 4096,
		Entries:      []BoardEntry{{Kills: 5, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date}},
		NewHighScore: true,
	}

	summary := toScoreSummary(request)

	assert.Equal(t, 3, summary.Kills)
	assert.Equal(t, 2, summary.Duds)
	assert.Equal(t, int64(4096), summary.FreedMem)
	assert.Equal(t, 7.5, summary.Duration)
	assert.True(t, summary.IsTopScore)
	assert.Equal(t, []outbound.ScoreEntry{{Kills: 5, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: date}}, summary.Entries)
}

func TestToScoreSummaryIsTopScoreFalseWhenNotANewHighScore(t *testing.T) {
	summary := toScoreSummary(ReportRequest{Kills: 5})

	assert.False(t, summary.IsTopScore)
}
