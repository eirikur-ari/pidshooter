package score

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

func toBoard(board outbound.ScoreBoard) *score.Board {
	entries := make([]score.Entry, len(board.Scores))
	for i, entry := range board.Scores {
		entries[i] = score.Entry{
			Kills:    entry.Kills,
			FreedMem: entry.FreedMem,
			Speed:    entry.Speed,
			Time:     entry.Time,
			Duration: entry.Duration,
			Date:     entry.Date,
		}
	}
	return score.NewBoard(entries)
}

func toScoreBoard(board *score.Board) outbound.ScoreBoard {
	entries := make([]outbound.ScoreEntry, len(board.Scores))
	for i, entry := range board.Scores {
		entries[i] = outbound.ScoreEntry{
			Kills:    entry.Kills,
			FreedMem: entry.FreedMem,
			Speed:    entry.Speed,
			Time:     entry.Time,
			Duration: entry.Duration,
			Date:     entry.Date,
		}
	}
	return outbound.ScoreBoard{Scores: entries}
}
