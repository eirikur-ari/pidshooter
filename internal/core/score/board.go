package score

import (
	"fmt"
	"sort"

	"github.com/eirikur-ari/pidshooter/internal/util"
)

const maxScores = 10

// Board holds all high scores and their ranking logic.
type Board struct {
	Scores    []Entry
	highScore int
}

// NewBoard builds a Board from persisted entries.
func NewBoard(entries []Entry) *Board {
	return &Board{Scores: entries}
}

// HighScore returns the board's current best kill count, or 0 if none.
func (b *Board) HighScore() int {
	return b.killScore()
}

// Add inserts a new score entry and keeps only the top N.
func (b *Board) Add(entry Entry) {
	b.highScore = b.killScore()
	b.Scores = append(b.Scores, entry)
	b.sortByRank()
	if len(b.Scores) > maxScores {
		b.Scores = b.Scores[:maxScores]
	}
}

// TODO: perhaps this should not be part of a domain layer, but rather a presentation layer concern, since it prints to stdout. Maybe move to a view package?
// PrintHighScores prints a trophy message if kills beats the high score
// recorded at the time of the last Add call, then displays the high score table.
func (b *Board) PrintHighScores(kills int) {
	if kills > 0 && kills >= b.highScore {
		fmt.Println("  🏆 New high score!")
	}

	if len(b.Scores) == 0 {
		fmt.Println("\n  No high scores yet!")
		return
	}

	fmt.Println("\n  ╔════╦═══════╦═══════╦════════╦════════════╦════════════╗")
	fmt.Println("  ║  # ║ Kills ║ Speed ║  Time  ║   Freed    ║    Date    ║")
	fmt.Println("  ╠════╬═══════╬═══════╬════════╬════════════╬════════════╣")

	for i, s := range b.Scores {
		mem := util.FormatBytes(s.FreedMem)
		date := s.Date.Format("2006-01-02")
		fmt.Printf("  ║ %2d ║  %3d  ║ %4.1fx ║ %5.1fs ║ %8s   ║ %s ║\n",
			i+1, s.Kills, s.Speed, s.Duration, mem, date)
	}

	fmt.Println("  ╚════╩═══════╩═══════╩════════╩════════════╩════════════╝")
}

// killScore returns the current top score (kills), or 0 if none.
func (b *Board) killScore() int {
	if len(b.Scores) == 0 {
		return 0
	}
	return b.Scores[0].Kills
}

func (b *Board) sortByRank() {
	sort.Slice(b.Scores, func(i, j int) bool {
		return b.Scores[i].beats(b.Scores[j])
	})
}
