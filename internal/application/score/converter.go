package score

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

func toBoard(sb outbound.ScoreBoard) *score.Board {
	entries := make([]score.Entry, len(sb.Scores))
	for i, e := range sb.Scores {
		entries[i] = score.Entry{
			Kills:    e.Kills,
			FreedMem: e.FreedMem,
			Speed:    e.Speed,
			Time:     e.Time,
			Duration: e.Duration,
			Date:     e.Date,
		}
	}
	return score.NewBoard(entries)
}

func toScoreBoard(b *score.Board) outbound.ScoreBoard {
	entries := make([]outbound.ScoreEntry, len(b.Scores))
	for i, e := range b.Scores {
		entries[i] = outbound.ScoreEntry{
			Kills:    e.Kills,
			FreedMem: e.FreedMem,
			Speed:    e.Speed,
			Time:     e.Time,
			Duration: e.Duration,
			Date:     e.Date,
		}
	}
	return outbound.ScoreBoard{Scores: entries}
}
