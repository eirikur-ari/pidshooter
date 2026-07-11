package score

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil/capture"
)

// --- Entry.beats ---

func TestEntry_Beats_ByKills(t *testing.T) {
	high := Entry{Kills: 10, FreedMem: 100}
	low := Entry{Kills: 5, FreedMem: 9000}
	assert.True(t, high.beats(low), "expected higher kills to win regardless of freed mem")
	assert.False(t, low.beats(high), "expected lower kills to lose")
}

func TestEntry_Beats_TiebreakByFreedMem(t *testing.T) {
	more := Entry{Kills: 5, FreedMem: 2000}
	less := Entry{Kills: 5, FreedMem: 1000}
	assert.True(t, more.beats(less), "expected higher freed mem to win on kills tie")
	assert.False(t, less.beats(more), "expected lower freed mem to lose on kills tie")
}

// --- Board.Add ---

func TestBoard_Add_SortsDescending(t *testing.T) {
	b := &Board{}

	b.Add(Entry{Kills: 3, FreedMem: 100, Date: time.Now()})
	b.Add(Entry{Kills: 7, FreedMem: 200, Date: time.Now()})
	b.Add(Entry{Kills: 5, FreedMem: 150, Date: time.Now()})

	require.Len(t, b.Scores, 3)
	assert.Equal(t, 7, b.Scores[0].Kills)
	assert.Equal(t, 5, b.Scores[1].Kills)
	assert.Equal(t, 3, b.Scores[2].Kills)
}

func TestBoard_Add_TiebreakByMemory(t *testing.T) {
	b := &Board{}

	b.Add(Entry{Kills: 5, FreedMem: 100, Date: time.Now()})
	b.Add(Entry{Kills: 5, FreedMem: 500, Date: time.Now()})

	assert.Equal(t, int64(500), b.Scores[0].FreedMem, "expected higher memory first")
}

func TestBoard_Add_CapsAtMax(t *testing.T) {
	b := &Board{}

	for i := 0; i < maxScores+5; i++ {
		b.Add(Entry{Kills: i, FreedMem: int64(i * 100), Date: time.Now()})
	}

	assert.Len(t, b.Scores, maxScores)

	expectedLowest := 5 // entries 0..4 are evicted; 5..14 are kept
	assert.Equal(t, expectedLowest, b.Scores[len(b.Scores)-1].Kills)
}

// --- Board.HighScore ---

func TestBoard_HighScore_Empty(t *testing.T) {
	b := &Board{}
	assert.Equal(t, 0, b.HighScore())
}

func TestBoard_HighScore_WithEntries(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 3, Date: time.Now()})
	b.Add(Entry{Kills: 10, Date: time.Now()})
	b.Add(Entry{Kills: 7, Date: time.Now()})

	assert.Equal(t, 10, b.HighScore())
}

// --- Board.PrintHighScore ---

func TestBoard_PrintHighScore_PrintsWhenBeatsRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 3, Date: time.Now()})
	b.Add(Entry{Kills: 7, Date: time.Now()}) // b.highScore = 3

	out := capture.Output(func() { b.PrintHighScore(7) })
	assert.Contains(t, out, "New high score")
}

func TestBoard_PrintHighScore_SilentWhenDoesNotBeatRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 7, Date: time.Now()})
	b.Add(Entry{Kills: 3, Date: time.Now()}) // b.highScore = 7

	out := capture.Output(func() { b.PrintHighScore(3) })
	assert.Empty(t, out)
}

func TestBoard_PrintHighScore_SilentWhenZeroKills(t *testing.T) {
	b := &Board{}
	out := capture.Output(func() { b.PrintHighScore(0) })
	assert.Empty(t, out)
}

func TestBoard_PrintHighScore_PrintsWhenTiesRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 5, Date: time.Now()}) // b.highScore = 0 before append
	b.Add(Entry{Kills: 3, Date: time.Now()}) // b.highScore = 5 before append

	out := capture.Output(func() { b.PrintHighScore(5) })
	assert.Contains(t, out, "New high score", "expected trophy message when tying the high score")
}

func TestBoard_PrintHighScore_PrintsForFirstEntry(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 5, Date: time.Now()}) // b.highScore = 0

	out := capture.Output(func() { b.PrintHighScore(5) })
	assert.Contains(t, out, "New high score", "expected trophy message for first entry")
}
