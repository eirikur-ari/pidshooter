package score

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

func TestToEntry_MapsRequestToEntry(t *testing.T) {
	// Given
	request := RecordRequest{Kills: 5, Duds: 2, FreedMem: 2048, LowestSpeed: 2.5, TimeLimit: 30, Duration: 12.5}

	// When
	entry := toEntry(request)

	// Then
	assert.Equal(t, 5, entry.Kills)
	assert.Equal(t, 2, entry.Duds)
	assert.Equal(t, int64(2048), entry.FreedMem)
	assert.Equal(t, 2.5, entry.Speed)
	assert.Equal(t, 30, entry.Time)
	assert.Equal(t, 12.5, entry.Duration)
	assert.False(t, entry.Date.IsZero())
}

func TestToEntries_MapsEveryEntryInOrder(t *testing.T) {
	tests := newToEntriesTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toEntries(test.entries)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToBoardEntries_MapsEveryEntryInOrder(t *testing.T) {
	tests := newToBoardEntriesTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toBoardEntries(test.entries)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToBoard_MapsStoredScoresToBoard(t *testing.T) {
	tests := newToBoardTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toBoard(test.board)

			// Then
			assert.Equal(t, test.expectedScores, actual.Scores)
			assert.Equal(t, test.expectedHighScore, actual.HighScore())
		})
	}
}

func TestToLoadResult_MapsStoredScoresToRankedEntriesAndHighScore(t *testing.T) {
	tests := newToLoadResultTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toLoadResult(test.board)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToRecordResult_MapsBoardAndFlagsNewHighScore(t *testing.T) {
	tests := newToRecordResultTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toRecordResult(test.board, test.kills)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToScoreBoard_MapsEntriesToStoredScores(t *testing.T) {
	tests := newToScoreBoardTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toScoreBoard(test.entries)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToScoreSummary_MapsReportRequest(t *testing.T) {
	tests := newToScoreSummaryTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toScoreSummary(test.request)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func newToEntriesTestCase() []struct {
	name     string
	entries  []BoardEntry
	expected []score.Entry
} {
	return []struct {
		name     string
		entries  []BoardEntry
		expected []score.Entry
	}{
		{
			name:     "several entries",
			entries:  []BoardEntry{newBoardEntryFixture(), {Kills: 3, Duds: 2}},
			expected: []score.Entry{newEntryFixture(), {Kills: 3, Duds: 2}},
		},
		{name: "no entries", entries: nil, expected: []score.Entry{}},
	}
}

func newToBoardEntriesTestCase() []struct {
	name     string
	entries  []score.Entry
	expected []BoardEntry
} {
	return []struct {
		name     string
		entries  []score.Entry
		expected []BoardEntry
	}{
		{
			name:     "several entries",
			entries:  []score.Entry{newEntryFixture(), {Kills: 3, Duds: 2}},
			expected: []BoardEntry{newBoardEntryFixture(), {Kills: 3, Duds: 2}},
		},
		{name: "no entries", entries: nil, expected: []BoardEntry{}},
	}
}

func newToBoardTestCase() []struct {
	name              string
	board             outbound.ScoreBoard
	expectedScores    []score.Entry
	expectedHighScore int
} {
	return []struct {
		name              string
		board             outbound.ScoreBoard
		expectedScores    []score.Entry
		expectedHighScore int
	}{
		{
			name:              "stored scores",
			board:             outbound.ScoreBoard{Scores: []outbound.ScoreEntry{newScoreEntryFixture()}},
			expectedScores:    []score.Entry{newEntryFixture()},
			expectedHighScore: 5,
		},
		{name: "no stored scores", board: outbound.ScoreBoard{}},
	}
}

func newToLoadResultTestCase() []struct {
	name     string
	board    outbound.ScoreBoard
	expected LoadResult
} {
	return []struct {
		name     string
		board    outbound.ScoreBoard
		expected LoadResult
	}{
		{
			name: "stored scores",
			board: outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
				{Kills: 3, Duds: 2, Date: dateFixture()},
				newScoreEntryFixture(),
			}},
			expected: LoadResult{
				Entries:   []BoardEntry{newBoardEntryFixture(), {Kills: 3, Duds: 2, Date: dateFixture()}},
				HighScore: 5,
			},
		},
		{name: "no stored scores", board: outbound.ScoreBoard{}, expected: LoadResult{Entries: []BoardEntry{}}},
	}
}

func newToRecordResultTestCase() []struct {
	name     string
	board    *score.Board
	kills    int
	expected RecordResult
} {
	return []struct {
		name     string
		board    *score.Board
		kills    int
		expected RecordResult
	}{
		{
			name:     "kills beat the best entry",
			board:    score.NewBoard([]score.Entry{newEntryFixture()}),
			kills:    6,
			expected: RecordResult{Entries: []BoardEntry{newBoardEntryFixture()}, NewHighScore: true},
		},
		{
			name:     "kills do not beat the best entry",
			board:    score.NewBoard([]score.Entry{newEntryFixture()}),
			kills:    5,
			expected: RecordResult{Entries: []BoardEntry{newBoardEntryFixture()}},
		},
		{
			name:     "empty board",
			board:    score.NewBoard(nil),
			kills:    0,
			expected: RecordResult{Entries: []BoardEntry{}},
		},
	}
}

func newToScoreBoardTestCase() []struct {
	name     string
	entries  []BoardEntry
	expected outbound.ScoreBoard
} {
	return []struct {
		name     string
		entries  []BoardEntry
		expected outbound.ScoreBoard
	}{
		{
			name:     "entries",
			entries:  []BoardEntry{newBoardEntryFixture()},
			expected: outbound.ScoreBoard{Scores: []outbound.ScoreEntry{newScoreEntryFixture()}},
		},
		{name: "no entries", entries: nil, expected: outbound.ScoreBoard{Scores: []outbound.ScoreEntry{}}},
	}
}

func newToScoreSummaryTestCase() []struct {
	name     string
	request  ReportRequest
	expected outbound.ScoreSummary
} {
	entries := []BoardEntry{newBoardEntryFixture()}
	expectedEntries := []outbound.ScoreEntry{newScoreEntryFixture()}

	return []struct {
		name     string
		request  ReportRequest
		expected outbound.ScoreSummary
	}{
		{
			name:    "new high score",
			request: ReportRequest{Duration: 7.5, Kills: 3, Duds: 2, FreedMem: 4096, Entries: entries, NewHighScore: true},
			expected: outbound.ScoreSummary{
				Kills: 3, Duds: 2, FreedMem: 4096, Duration: 7.5, IsTopScore: true, Entries: expectedEntries,
			},
		},
		{
			name:    "not a new high score",
			request: ReportRequest{Duration: 7.5, Kills: 3, Duds: 2, FreedMem: 4096, Entries: entries},
			expected: outbound.ScoreSummary{
				Kills: 3, Duds: 2, FreedMem: 4096, Duration: 7.5, Entries: expectedEntries,
			},
		},
	}
}
