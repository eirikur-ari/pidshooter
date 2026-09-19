package score

import (
	"sort"
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

// IsNewHighScore reports whether kills would have beaten the board's
// high score as of the last Add call.
func (b *Board) IsNewHighScore(kills int) bool {
	return kills > 0 && kills >= b.highScore
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
