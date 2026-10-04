package score

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBoard_AppliesTheSameRulesAsAdd(t *testing.T) {
	// Given
	noKills := newEntryFixtureFor(0)
	negativeDuds := newEntryFixture()
	negativeDuds.Duds = -1
	entries := []Entry{noKills}
	for kills := 1; kills <= maxScores+5; kills++ { // ascending kills: deliberately unsorted
		entries = append(entries, newEntryFixtureFor(kills))
	}
	entries = append(entries, negativeDuds)
	expected := &Board{}
	for _, entry := range entries {
		expected.Add(entry)
	}

	// When
	board := NewBoard(entries)

	// Then
	require.Len(t, board.Scores, maxScores, "the input must have been both filtered and capped")
	assert.Equal(t, expected.Scores, board.Scores)
}

func TestNewBoard_SeedsHighScoreFromEntries(t *testing.T) {
	tests := []struct {
		name  string
		kills []int
	}{
		{name: "no entries", kills: nil},
		{name: "single entry", kills: []int{5}},
		{name: "high score entry first", kills: []int{7, 3}},
		{name: "high score entry last", kills: []int{3, 7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			var entries []Entry
			for _, kills := range tt.kills {
				entries = append(entries, newEntryFixtureFor(kills))
			}
			board := NewBoard(entries)
			highScore := board.HighScore()

			// Then
			assert.False(t, board.IsNewHighScore(highScore-1), "fewer kills than the current high score entry is not a new high score")
			assert.False(t, board.IsNewHighScore(highScore), "tying the current high score entry is not a new high score")
			assert.True(t, board.IsNewHighScore(highScore+1), "more kills than the current high score entry is a new high score")
		})
	}
}

func TestBoard_Add_RanksEntriesBestFirst(t *testing.T) {
	// Given
	board := &Board{}
	three := newEntryFixtureFor(3)
	seven := newEntryFixtureFor(7)
	fiveLessMemory := newEntryFixtureFor(5)
	fiveLessMemory.FreedMem = 100
	fiveMoreMemory := newEntryFixtureFor(5)
	fiveMoreMemory.FreedMem = 500

	// When
	for _, entry := range []Entry{three, seven, fiveLessMemory, fiveMoreMemory} {
		board.Add(entry)
	}

	// Then
	assert.Equal(t, []Entry{seven, fiveMoreMemory, fiveLessMemory, three}, board.Scores)
}

func TestBoard_Add_KeepsInsertionOrderWhenAllRankedFieldsTie(t *testing.T) {
	// Given
	board := &Board{}
	// Date isn't one of beats ranked dimensions, so it's a safe marker for
	// telling otherwise-identical entries apart without affecting rank.
	tied := func(marker int) Entry {
		entry := newEntryFixture()
		entry.Date = time.Unix(int64(marker), 0)
		return entry
	}

	// When
	for i := 1; i <= 5; i++ {
		board.Add(tied(i))
	}

	// Then
	require.Len(t, board.Scores, 5)
	for i, entry := range board.Scores {
		assert.Equal(t, i+1, int(entry.Date.Unix()), "a full tie across every ranked field must deterministically preserve insertion order, not depend on sort.Slice's unspecified tie-breaking")
	}
}

func TestBoard_Add_KeepsOnlyMaxNumberOfHighScores(t *testing.T) {
	// Given
	const (
		extraEntries        = 5
		expectedTopScore    = maxScores + extraEntries
		expectedLowestScore = extraEntries + 1
	)
	board := &Board{}

	// When
	for kills := 1; kills <= maxScores+extraEntries; kills++ {
		board.Add(newEntryFixtureFor(kills))
	}
	topScore := board.Scores[0].Kills
	lowestScore := board.Scores[len(board.Scores)-1].Kills

	// Then
	require.Len(t, board.Scores, maxScores)
	assert.Equal(t, expectedTopScore, topScore)
	assert.Equal(t, expectedLowestScore, lowestScore, "the lowest-ranked extra entries must be evicted")
}

func TestBoard_Add_IgnoresEntryThatIsNotAGenuineScore(t *testing.T) {
	// Given
	board := &Board{}

	// When
	board.Add(newEntryFixtureFor(0))

	// Then
	assert.Empty(t, board.Scores)
}

func TestBoard_Add_LeavesHighScoreUnchangedWhenEntryIsIgnored(t *testing.T) {
	// Given
	board := &Board{}

	// When
	board.Add(newEntryFixtureFor(3))
	board.Add(newEntryFixtureFor(7))
	board.Add(newEntryFixtureFor(0))

	// Then
	assert.True(t, board.IsNewHighScore(5), "an ignored entry must not move the previous high score from 3 to 7")
}

func TestBoard_Add_RaisesOnlyNegativeSpeedAndDurationToZero(t *testing.T) {
	tests := []struct {
		name             string
		speed            float64
		duration         float64
		expectedSpeed    float64
		expectedDuration float64
	}{
		{name: "both negative", speed: -1.0, duration: -5.0, expectedSpeed: 0, expectedDuration: 0},
		{name: "negative speed keeps positive duration", speed: -1.0, duration: 10.0, expectedSpeed: 0, expectedDuration: 10.0},
		{name: "negative duration keeps positive speed", speed: 0.5, duration: -1.0, expectedSpeed: 0.5, expectedDuration: 0},
		{name: "both positive are preserved", speed: 2.0, duration: 10.0, expectedSpeed: 2.0, expectedDuration: 10.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			board := &Board{}
			entry := newEntryFixture()
			entry.Speed = tt.speed
			entry.Duration = tt.duration

			// When
			board.Add(entry)

			// Then
			require.Len(t, board.Scores, 1)
			assert.Equal(t, tt.expectedSpeed, board.Scores[0].Speed)
			assert.Equal(t, tt.expectedDuration, board.Scores[0].Duration)
		})
	}
}

func TestBoard_HighScore_IsZeroWhenEmpty(t *testing.T) {
	// Given
	board := &Board{}

	// Then
	assert.Equal(t, 0, board.HighScore())
}

func TestBoard_HighScore_IsTopEntryKills(t *testing.T) {
	// Given
	board := &Board{}

	// When
	board.Add(newEntryFixtureFor(3))
	board.Add(newEntryFixtureFor(10))
	board.Add(newEntryFixtureFor(7))

	// Then
	assert.Equal(t, 10, board.HighScore())
}

func TestBoard_IsNewHighScore_ComparesAgainstPreviousHighScore(t *testing.T) {
	tests := []struct {
		name       string
		addedKills []int
		kills      int
		expected   bool
	}{
		{name: "first entry on an empty board", addedKills: []int{5}, kills: 5, expected: true},
		{name: "beats the previous high score", addedKills: []int{3, 7}, kills: 7, expected: true},
		{name: "below the previous high score", addedKills: []int{7, 3}, kills: 3, expected: false},
		{name: "ties the previous high score", addedKills: []int{5, 3}, kills: 5, expected: false},
		{name: "zero kills", addedKills: nil, kills: 0, expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			board := &Board{}
			for _, kills := range tt.addedKills {
				board.Add(newEntryFixtureFor(kills))
			}

			// When
			isNew := board.IsNewHighScore(tt.kills)

			// Then
			assert.Equal(t, tt.expected, isNew)
		})
	}
}
