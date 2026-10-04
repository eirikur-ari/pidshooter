package score

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Board.Add ---

func TestBoardAddSortsDescending(t *testing.T) {
	b := &Board{}

	b.Add(Entry{Kills: 3, FreedMem: 100, Date: time.Now()})
	b.Add(Entry{Kills: 7, FreedMem: 200, Date: time.Now()})
	b.Add(Entry{Kills: 5, FreedMem: 150, Date: time.Now()})

	require.Len(t, b.Scores, 3)
	assert.Equal(t, 7, b.Scores[0].Kills)
	assert.Equal(t, 5, b.Scores[1].Kills)
	assert.Equal(t, 3, b.Scores[2].Kills)
}

func TestBoardAddTiebreakByMemory(t *testing.T) {
	b := &Board{}

	b.Add(Entry{Kills: 5, FreedMem: 100, Date: time.Now()})
	b.Add(Entry{Kills: 5, FreedMem: 500, Date: time.Now()})

	assert.Equal(t, int64(500), b.Scores[0].FreedMem, "expected higher memory first")
}

func TestBoardAddIgnoresZeroKillEntry(t *testing.T) {
	b := &Board{}

	b.Add(Entry{Kills: 0, Date: time.Now()})

	assert.Empty(t, b.Scores, "a zero-kill session is not a score and should not be recorded")
}

func TestBoardAddIgnoresNegativeKillEntry(t *testing.T) {
	b := &Board{}

	b.Add(Entry{Kills: -1, Date: time.Now()})

	assert.Empty(t, b.Scores)
}

func TestBoardAddIgnoresNegativeFreedMemEntry(t *testing.T) {
	b := &Board{}

	b.Add(Entry{Kills: 5, FreedMem: -1, Date: time.Now()})

	assert.Empty(t, b.Scores, "negative freed memory cannot come from a real session and should not be recorded")
}

func TestBoardAddFullTieKeepsInsertionOrder(t *testing.T) {
	b := &Board{}
	// Date isn't one of beats's ranked dimensions, so it's a safe marker for
	// telling otherwise-identical entries apart without affecting rank.
	tied := func(marker int) Entry {
		return Entry{Kills: 5, Duds: 1, Speed: 2.0, Duration: 10.0, FreedMem: 1024, Date: time.Unix(int64(marker), 0)}
	}

	for i := 1; i <= 5; i++ {
		b.Add(tied(i))
	}

	require.Len(t, b.Scores, 5)
	for i, entry := range b.Scores {
		assert.Equal(t, i+1, int(entry.Date.Unix()), "a full tie across every ranked field must deterministically preserve insertion order, not depend on sort.Slice's unspecified tie-breaking")
	}
}

func TestBoardAddCapsAtMax(t *testing.T) {
	b := &Board{}

	for i := range maxScores + 5 {
		b.Add(Entry{Kills: i, FreedMem: int64(i * 100), Date: time.Now()})
	}

	assert.Len(t, b.Scores, maxScores)

	expectedLowest := 5 // entries 0..4 are evicted; 5..14 are kept
	assert.Equal(t, expectedLowest, b.Scores[len(b.Scores)-1].Kills)
}

// --- Board.killScore ---

func TestBoardHighScoreEmpty(t *testing.T) {
	b := &Board{}
	assert.Equal(t, 0, b.killScore())
}

func TestBoardHighScoreWithEntries(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 3, Date: time.Now()})
	b.Add(Entry{Kills: 10, Date: time.Now()})
	b.Add(Entry{Kills: 7, Date: time.Now()})

	assert.Equal(t, 10, b.killScore())
}

// --- NewBoard ---

func TestNewBoardPopulatesEntries(t *testing.T) {
	entries := []Entry{{Kills: 5, Date: time.Now()}}

	b := NewBoard(entries)

	assert.Equal(t, entries, b.Scores)
	assert.Equal(t, 5, b.HighScore())
}

func TestNewBoardSeedsHighScoreFromEntries(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
	}{
		{name: "single entry", entries: []Entry{{Kills: 5}}},
		{name: "high score entry first", entries: []Entry{{Kills: 7}, {Kills: 3}}},
		{name: "high score entry last", entries: []Entry{{Kills: 3}, {Kills: 7}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			b := NewBoard(tt.entries)
			highScore := b.HighScore()

			// Then
			assert.False(t, b.IsNewHighScore(highScore-1), "fewer kills than the current high score entry is not a new high score")
			assert.False(t, b.IsNewHighScore(highScore), "tying the current high score entry is not a new high score")
			assert.True(t, b.IsNewHighScore(highScore+1), "more kills than the current high score entry is a new high score")
		})
	}
}

func TestNewBoardEmptyEntriesSeedsZeroHighScore(t *testing.T) {
	b := NewBoard(nil)

	assert.Empty(t, b.Scores)
	assert.Equal(t, 0, b.HighScore())
}

func TestNewBoardDropsEntriesThatAreNotGenuineScores(t *testing.T) {
	entries := []Entry{
		{Kills: 5, FreedMem: 100, Date: time.Now()},
		{Kills: 0, FreedMem: 100, Date: time.Now()},  // no kills
		{Kills: -1, FreedMem: 100, Date: time.Now()}, // negative kills
		{Kills: 5, FreedMem: -1, Date: time.Now()},   // negative freed mem
	}

	b := NewBoard(entries)

	require.Len(t, b.Scores, 1, "entries that aren't genuine scores must be dropped")
	assert.Equal(t, 100, int(b.Scores[0].FreedMem))
}

func TestNewBoardSortsAndCapsOversizedUnsortedInput(t *testing.T) {
	entries := make([]Entry, 0, maxScores+5)
	for i := maxScores + 4; i >= 0; i-- { // deliberately unsorted: ascending Kills
		entries = append(entries, Entry{Kills: i, Date: time.Now()})
	}

	b := NewBoard(entries)

	require.Len(t, b.Scores, maxScores, "more than maxScores entries must be capped")
	assert.Equal(t, maxScores+4, b.Scores[0].Kills, "entries must be ranked, not trusted to arrive in order")
	assert.Equal(t, 5, b.Scores[len(b.Scores)-1].Kills)
}

// --- Board.IsNewHighScore ---

func TestBoardIsNewHighScoreTrueWhenBeatsRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 3, Date: time.Now()})
	b.Add(Entry{Kills: 7, Date: time.Now()}) // b.highScore = 3

	assert.True(t, b.IsNewHighScore(7))
}

func TestBoardIsNewHighScoreFalseWhenDoesNotBeatRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 7, Date: time.Now()})
	b.Add(Entry{Kills: 3, Date: time.Now()}) // b.highScore = 7

	assert.False(t, b.IsNewHighScore(3))
}

func TestBoardIsNewHighScoreFalseWhenZeroKills(t *testing.T) {
	b := &Board{}
	assert.False(t, b.IsNewHighScore(0))
}

func TestBoardIsNewHighScoreFalseWhenTiesRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 5, Date: time.Now()}) // b.highScore = 0 before append
	b.Add(Entry{Kills: 3, Date: time.Now()}) // b.highScore = 5 before append

	assert.False(t, b.IsNewHighScore(5), "tying the high score is not a new high score")
}

func TestBoardIsNewHighScoreTrueForFirstEntry(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 5, Date: time.Now()}) // b.highScore = 0

	assert.True(t, b.IsNewHighScore(5), "expected true for first entry")
}
