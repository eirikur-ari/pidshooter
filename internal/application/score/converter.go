package score

import (
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

func ToEntry(result game.PlayResult, cfgTimeLimit int) score.Entry {
	return score.Entry{
		Kills:    result.Kills,
		Duds:     len(result.Duds),
		FreedMem: result.FreedMem,
		Speed:    result.LowestSpeed,
		Time:     cfgTimeLimit,
		Duration: result.Duration,
		Date:     time.Now(),
	}
}

func toBoard(board outbound.ScoreBoard) *score.Board {
	entries := make([]score.Entry, len(board.Scores))
	for i, entry := range board.Scores {
		entries[i] = score.Entry{
			Kills:    entry.Kills,
			Duds:     entry.Duds,
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
	return outbound.ScoreBoard{Scores: toScoreEntries(board.Scores)}
}

func toScoreSummary(duration float64, kills, duds int, freedMem int64, board *score.Board) outbound.ScoreSummary {
	return outbound.ScoreSummary{
		Kills:        kills,
		Duds:         duds,
		FreedMem:     freedMem,
		Duration:     duration,
		NewHighScore: board.IsNewHighScore(kills),
		Entries:      toScoreEntries(board.Scores),
	}
}

func toScoreEntries(entries []score.Entry) []outbound.ScoreEntry {
	result := make([]outbound.ScoreEntry, len(entries))
	for i, entry := range entries {
		result[i] = outbound.ScoreEntry{
			Kills:    entry.Kills,
			Duds:     entry.Duds,
			FreedMem: entry.FreedMem,
			Speed:    entry.Speed,
			Time:     entry.Time,
			Duration: entry.Duration,
			Date:     entry.Date,
		}
	}
	return result
}
