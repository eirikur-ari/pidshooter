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

func TestNewBoardEmptyEntriesSeedsZeroHighScore(t *testing.T) {
	b := NewBoard(nil)

	assert.Empty(t, b.Scores)
	assert.Equal(t, 0, b.HighScore())
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

func TestBoardIsNewHighScoreTrueWhenTiesRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 5, Date: time.Now()}) // b.highScore = 0 before append
	b.Add(Entry{Kills: 3, Date: time.Now()}) // b.highScore = 5 before append

	assert.True(t, b.IsNewHighScore(5), "expected true when tying the high score")
}

func TestBoardIsNewHighScoreTrueForFirstEntry(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 5, Date: time.Now()}) // b.highScore = 0

	assert.True(t, b.IsNewHighScore(5), "expected true for first entry")
}
