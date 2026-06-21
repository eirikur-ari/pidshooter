package score

import (
	"testing"
	"time"
)

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

	lowestKills := b.Scores[len(b.Scores)-1].Kills
	if lowestKills < 5 {
		t.Errorf("expected lowest retained score >= 5, got %d", lowestKills)
	}
}

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
