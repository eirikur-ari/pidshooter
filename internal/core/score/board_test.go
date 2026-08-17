package score

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil/capture"
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

	for i := 0; i < maxScores+5; i++ {
		b.Add(Entry{Kills: i, FreedMem: int64(i * 100), Date: time.Now()})
	}

	assert.Len(t, b.Scores, maxScores)

	expectedLowest := 5 // entries 0..4 are evicted; 5..14 are kept
	assert.Equal(t, expectedLowest, b.Scores[len(b.Scores)-1].Kills)
}

// --- Board.HighScore ---

func TestBoardHighScoreEmpty(t *testing.T) {
	b := &Board{}
	assert.Equal(t, 0, b.HighScore())
}

func TestBoardHighScoreWithEntries(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 3, Date: time.Now()})
	b.Add(Entry{Kills: 10, Date: time.Now()})
	b.Add(Entry{Kills: 7, Date: time.Now()})

	assert.Equal(t, 10, b.HighScore())
}

// --- NewBoard ---

func TestNewBoardPopulatesEntries(t *testing.T) {
	entries := []Entry{{Kills: 5, Date: time.Now()}}

	b, tracker := NewBoard(entries)

	assert.Equal(t, entries, b.Scores)
	assert.Equal(t, 5, tracker.HighScore)
}

func TestNewBoardEmptyEntriesSeedsZeroHighScore(t *testing.T) {
	b, tracker := NewBoard(nil)

	assert.Empty(t, b.Scores)
	assert.Equal(t, 0, tracker.HighScore)
}

// --- Board.PrintHighScores ---

func TestBoardPrintScoresPrintsTrophyWhenBeatsRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 3, Date: time.Now()})
	b.Add(Entry{Kills: 7, Date: time.Now()}) // b.highScore = 3
	b.tracker = &Tracker{Kills: 7}

	out := capture.Output(func() { b.PrintHighScores() })
	assert.Contains(t, out, "New high score")
}

func TestBoardPrintScoresNoTrophyWhenDoesNotBeatRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 7, Date: time.Now()})
	b.Add(Entry{Kills: 3, Date: time.Now()}) // b.highScore = 7
	b.tracker = &Tracker{Kills: 3}

	out := capture.Output(func() { b.PrintHighScores() })
	assert.NotContains(t, out, "New high score")
}

func TestBoardPrintScoresNoTrophyWhenZeroKills(t *testing.T) {
	b := &Board{tracker: &Tracker{}}
	out := capture.Output(func() { b.PrintHighScores() })
	assert.NotContains(t, out, "New high score")
}

func TestBoardPrintScoresPrintsTrophyWhenTiesRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 5, Date: time.Now()}) // b.highScore = 0 before append
	b.Add(Entry{Kills: 3, Date: time.Now()}) // b.highScore = 5 before append
	b.tracker = &Tracker{Kills: 5}

	out := capture.Output(func() { b.PrintHighScores() })
	assert.Contains(t, out, "New high score", "expected trophy message when tying the high score")
}

func TestBoardPrintScoresPrintsTrophyForFirstEntry(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 5, Date: time.Now()}) // b.highScore = 0
	b.tracker = &Tracker{Kills: 5}

	out := capture.Output(func() { b.PrintHighScores() })
	assert.Contains(t, out, "New high score", "expected trophy message for first entry")
}

func TestBoardPrintScoresPrintsNoScoresMessageWhenEmpty(t *testing.T) {
	b := &Board{tracker: &Tracker{}}
	out := capture.Output(func() { b.PrintHighScores() })
	assert.Contains(t, out, "No high scores yet!")
}

func TestBoardPrintScoresPrintsTable(t *testing.T) {
	b := &Board{tracker: &Tracker{}}
	b.Add(Entry{Kills: 5, FreedMem: 100, Speed: 1.5, Date: time.Now()})

	out := capture.Output(func() { b.PrintHighScores() })
	assert.Contains(t, out, "Kills")
	assert.Contains(t, out, "Freed")
}

func TestBoardPrintScoresPrintsDuration(t *testing.T) {
	b := &Board{tracker: &Tracker{}}
	b.Add(Entry{Kills: 5, FreedMem: 100, Speed: 1.5, Duration: 12.3, Date: time.Now()})

	out := capture.Output(func() { b.PrintHighScores() })
	assert.Contains(t, out, "Time")
	assert.Contains(t, out, "12.3s")
}
