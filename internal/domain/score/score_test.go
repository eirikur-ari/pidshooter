package score

import (
	"strings"
	"testing"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/testutil/capture"
)

// --- Entry.beats ---

func TestEntry_Beats_ByKills(t *testing.T) {
	high := Entry{Kills: 10, FreedMem: 100}
	low := Entry{Kills: 5, FreedMem: 9000}
	if !high.beats(low) {
		t.Error("expected higher kills to win regardless of freed mem")
	}
	if low.beats(high) {
		t.Error("expected lower kills to lose")
	}
}

func TestEntry_Beats_TiebreakByFreedMem(t *testing.T) {
	more := Entry{Kills: 5, FreedMem: 2000}
	less := Entry{Kills: 5, FreedMem: 1000}
	if !more.beats(less) {
		t.Error("expected higher freed mem to win on kills tie")
	}
	if less.beats(more) {
		t.Error("expected lower freed mem to lose on kills tie")
	}
}

// --- Board.Add ---

func TestBoard_Add_SortsDescending(t *testing.T) {
	b := &Board{}

	b.Add(Entry{Kills: 3, FreedMem: 100, Date: time.Now()})
	b.Add(Entry{Kills: 7, FreedMem: 200, Date: time.Now()})
	b.Add(Entry{Kills: 5, FreedMem: 150, Date: time.Now()})

	if len(b.Scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(b.Scores))
	}
	if b.Scores[0].Kills != 7 {
		t.Errorf("expected top score=7, got %d", b.Scores[0].Kills)
	}
	if b.Scores[1].Kills != 5 {
		t.Errorf("expected second score=5, got %d", b.Scores[1].Kills)
	}
	if b.Scores[2].Kills != 3 {
		t.Errorf("expected third score=3, got %d", b.Scores[2].Kills)
	}
}

func TestBoard_Add_TiebreakByMemory(t *testing.T) {
	b := &Board{}

	b.Add(Entry{Kills: 5, FreedMem: 100, Date: time.Now()})
	b.Add(Entry{Kills: 5, FreedMem: 500, Date: time.Now()})

	if b.Scores[0].FreedMem != 500 {
		t.Errorf("expected higher memory first, got %d", b.Scores[0].FreedMem)
	}
}

func TestBoard_Add_CapsAtMax(t *testing.T) {
	b := &Board{}

	for i := 0; i < maxScores+5; i++ {
		b.Add(Entry{Kills: i, FreedMem: int64(i * 100), Date: time.Now()})
	}

	if len(b.Scores) != maxScores {
		t.Errorf("expected %d scores, got %d", maxScores, len(b.Scores))
	}

	expectedLowest := 5 // entries 0..4 are evicted; 5..14 are kept
	if got := b.Scores[len(b.Scores)-1].Kills; got != expectedLowest {
		t.Errorf("expected lowest retained kills=%d, got %d", expectedLowest, got)
	}
}

// --- Board.HighScore ---

func TestBoard_HighScore_Empty(t *testing.T) {
	b := &Board{}
	if b.HighScore() != 0 {
		t.Errorf("expected 0 for empty board, got %d", b.HighScore())
	}
}

func TestBoard_HighScore_WithEntries(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 3, Date: time.Now()})
	b.Add(Entry{Kills: 10, Date: time.Now()})
	b.Add(Entry{Kills: 7, Date: time.Now()})

	if b.HighScore() != 10 {
		t.Errorf("expected high score=10, got %d", b.HighScore())
	}
}

// --- Board.PrintHighScore ---

func TestBoard_PrintHighScore_PrintsWhenBeatsRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 3, Date: time.Now()})
	b.Add(Entry{Kills: 7, Date: time.Now()}) // b.highScore = 3

	out := capture.Output(func() { b.PrintHighScore(7) })
	if !strings.Contains(out, "New high score") {
		t.Errorf("expected trophy message, got %q", out)
	}
}

func TestBoard_PrintHighScore_SilentWhenDoesNotBeatRecord(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 7, Date: time.Now()})
	b.Add(Entry{Kills: 3, Date: time.Now()}) // b.highScore = 7

	out := capture.Output(func() { b.PrintHighScore(3) })
	if out != "" {
		t.Errorf("expected no output, got %q", out)
	}
}

func TestBoard_PrintHighScore_SilentWhenZeroKills(t *testing.T) {
	b := &Board{}
	out := capture.Output(func() { b.PrintHighScore(0) })
	if out != "" {
		t.Errorf("expected no output for zero kills, got %q", out)
	}
}

func TestBoard_PrintHighScore_PrintsForFirstEntry(t *testing.T) {
	b := &Board{}
	b.Add(Entry{Kills: 5, Date: time.Now()}) // b.highScore = 0

	out := capture.Output(func() { b.PrintHighScore(5) })
	if !strings.Contains(out, "New high score") {
		t.Errorf("expected trophy message for first entry, got %q", out)
	}
}
