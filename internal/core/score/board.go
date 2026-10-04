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

// NewBoard builds a Board from entries, ignoring any that aren't genuine
// scores. The board's high score starts at its top entry's kills.
func NewBoard(entries []Entry) *Board {
	b := &Board{}
	for _, entry := range entries {
		b.insert(entry)
	}
	b.highScore = b.killScore()
	return b
}

// HighScore returns the board's current best kill count, or 0 if none.
func (b *Board) HighScore() int {
	return b.killScore()
}

// Add inserts a new score entry and keeps only the top N. An entry that
// isn't a genuine score is silently ignored.
func (b *Board) Add(entry Entry) {
	topScore := b.killScore()
	if b.insert(entry) {
		b.highScore = topScore
	}
}

// IsNewHighScore reports whether kills is strictly higher than the board's
// previous high score.
func (b *Board) IsNewHighScore(kills int) bool {
	return kills > 0 && kills > b.highScore
}

// killScore returns the current top score (kills), or 0 if none.
func (b *Board) killScore() int {
	if len(b.Scores) == 0 {
		return 0
	}

	return b.Scores[0].Kills
}

// insert ranks entry on the board and keeps only the top N, reporting whether
// it was stored. An entry that isn't a genuine score is ignored.
func (b *Board) insert(entry Entry) bool {
	if !entry.isScore() {
		return false
	}

	b.Scores = append(b.Scores, entry)
	b.sortByRank()

	if len(b.Scores) > maxScores {
		b.Scores = b.Scores[:maxScores]
	}
	return true
}

func (b *Board) sortByRank() {
	sort.SliceStable(b.Scores, b.less)
}

// less reports whether the entry at index i outranks the entry at index j.
func (b *Board) less(i, j int) bool {
	return b.Scores[i].beats(b.Scores[j])
}
