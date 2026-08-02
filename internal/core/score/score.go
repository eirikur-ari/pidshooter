// Package score manages high-score ranking logic.
package score

import (
	"fmt"
	"sort"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/util"
)

// Entry represents a single high score record.
type Entry struct {
	Kills    int
	FreedMem int64
	Speed    float64
	Time     int
	Duration float64
	Date     time.Time
}

// Board holds all high scores and their ranking logic.
type Board struct {
	Scores    []Entry
	highScore int
}

const maxScores = 10

// Add inserts a new score entry and keeps only the top N.
func (b *Board) Add(entry Entry) {
	b.highScore = b.HighScore()
	b.Scores = append(b.Scores, entry)
	b.sortByRank()
	if len(b.Scores) > maxScores {
		b.Scores = b.Scores[:maxScores]
	}
}

// HighScore returns the current top score (kills), or 0 if none.
func (b *Board) HighScore() int {
	if len(b.Scores) == 0 {
		return 0
	}
	return b.Scores[0].Kills
}

// PrintHighScore prints a trophy message if kills beats the high score
// recorded at the time of the last Add call.
func (b *Board) PrintHighScore(kills int) {
	if kills > 0 && kills >= b.highScore {
		fmt.Println("  🏆 New high score!")
	}
}

// PrintScores displays the high score table to stdout.
func (b *Board) PrintScores() {
	if len(b.Scores) == 0 {
		fmt.Println("\n  No high scores yet!")
		return
	}

	//TODO: Replace Speed with Time
	fmt.Println("\n  ╔════╦═══════╦════════════╦═══════╦════════════╗")
	fmt.Println("  ║  # ║ Kills ║   Freed    ║ Speed ║    Date    ║")
	fmt.Println("  ╠════╬═══════╬════════════╬═══════╬════════════╣")

	for i, s := range b.Scores {
		mem := util.FormatBytes(s.FreedMem)
		date := s.Date.Format("2006-01-02")
		fmt.Printf("  ║ %2d ║  %3d  ║ %8s   ║ %4.1fx ║ %s ║\n",
			i+1, s.Kills, mem, s.Speed, date)
	}

	fmt.Println("  ╚════╩═══════╩════════════╩═══════╩════════════╝")
}

func (b *Board) sortByRank() {
	sort.Slice(b.Scores, func(i, j int) bool {
		return b.Scores[i].beats(b.Scores[j])
	})
}

func (e Entry) beats(other Entry) bool {
	if e.Kills != other.Kills {
		return e.Kills > other.Kills
	}
	return e.FreedMem > other.FreedMem
}
