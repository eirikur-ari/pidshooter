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
	tests := newSeededHighScoreTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			var entries []Entry
			for _, kills := range test.kills {
				entries = append(entries, newEntryFixtureFor(kills))
			}

			// When
			board := NewBoard(entries)
			highScore := board.HighScore()
			below := board.IsNewHighScore(highScore - 1)
			tie := board.IsNewHighScore(highScore)
			above := board.IsNewHighScore(highScore + 1)

			// Then
			assert.False(t, below, "fewer kills than the current high score entry is not a new high score")
			assert.False(t, tie, "tying the current high score entry is not a new high score")
			assert.True(t, above, "more kills than the current high score entry is a new high score")
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
	expectedMarkers := []int{1, 2, 3, 4, 5}

	// When
	for _, marker := range expectedMarkers {
		board.Add(tied(marker))
	}
	var markers []int
	for _, entry := range board.Scores {
		markers = append(markers, int(entry.Date.Unix()))
	}

	// Then
	require.Len(t, board.Scores, 5)
	assert.Equal(t, expectedMarkers, markers, "a full tie across every ranked field must deterministically preserve insertion order, not depend on sort.Slice's unspecified tie-breaking")
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
	isHighScore := board.IsNewHighScore(5)

	// Then
	assert.True(t, isHighScore, "an ignored entry must not move the previous high score from 3 to 7")
}

func TestBoard_Add_RaisesOnlyNegativeSpeedAndDurationToZero(t *testing.T) {
	tests := newSpeedAndDurationTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			board := &Board{}
			entry := newEntryFixture()
			entry.Speed = test.speed
			entry.Duration = test.duration

			// When
			board.Add(entry)

			// Then
			require.Len(t, board.Scores, 1)
			assert.Equal(t, test.expectedSpeed, board.Scores[0].Speed)
			assert.Equal(t, test.expectedDuration, board.Scores[0].Duration)
		})
	}
}

func TestBoard_HighScore_IsZeroWhenEmpty(t *testing.T) {
	// Given
	board := &Board{}

	// When
	highScore := board.HighScore()

	// Then
	assert.Equal(t, 0, highScore)
}

func TestBoard_HighScore_IsTopEntryKills(t *testing.T) {
	// Given
	board := &Board{}

	// When
	board.Add(newEntryFixtureFor(3))
	board.Add(newEntryFixtureFor(10))
	board.Add(newEntryFixtureFor(7))
	highScore := board.HighScore()

	// Then
	assert.Equal(t, 10, highScore)
}

func TestBoard_IsNewHighScore_ComparesAgainstPreviousHighScore(t *testing.T) {
	tests := newNewHighScoreTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			board := &Board{}
			for _, kills := range test.addedKills {
				board.Add(newEntryFixtureFor(kills))
			}

			// When
			isNew := board.IsNewHighScore(test.kills)

			// Then
			assert.Equal(t, test.expected, isNew)
		})
	}
}

func newSeededHighScoreTestCase() []struct {
	name  string
	kills []int
} {
	return []struct {
		name  string
		kills []int
	}{
		{"no entries", nil},
		{"single entry", []int{5}},
		{"high score entry first", []int{7, 3}},
		{"high score entry last", []int{3, 7}},
	}
}

func newSpeedAndDurationTestCase() []struct {
	name             string
	speed            float64
	duration         float64
	expectedSpeed    float64
	expectedDuration float64
} {
	return []struct {
		name             string
		speed            float64
		duration         float64
		expectedSpeed    float64
		expectedDuration float64
	}{
		{"both negative", -1.0, -5.0, 0, 0},
		{"negative speed keeps positive duration", -1.0, 10.0, 0, 10.0},
		{"negative duration keeps positive speed", 0.5, -1.0, 0.5, 0},
		{"both positive are preserved", 2.0, 10.0, 2.0, 10.0},
	}
}

func newNewHighScoreTestCase() []struct {
	name       string
	addedKills []int
	kills      int
	expected   bool
} {
	return []struct {
		name       string
		addedKills []int
		kills      int
		expected   bool
	}{
		{"first entry on an empty board", []int{5}, 5, true},
		{"beats the previous high score", []int{3, 7}, 7, true},
		{"below the previous high score", []int{7, 3}, 3, false},
		{"ties the previous high score", []int{5, 3}, 5, false},
		{"zero kills", nil, 0, false},
	}
}
