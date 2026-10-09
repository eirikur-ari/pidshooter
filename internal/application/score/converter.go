package score

import (
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

func toEntry(request RecordRequest) score.Entry {
	return score.Entry{
		Kills:    request.Kills,
		Duds:     request.Duds,
		FreedMem: request.FreedMem,
		Speed:    request.LowestSpeed,
		Time:     request.TimeLimit,
		Duration: request.Duration,
		Date:     time.Now(),
	}
}

func toEntries(entries []BoardEntry) []score.Entry {
	out := make([]score.Entry, len(entries))
	for i, entry := range entries {
		out[i] = score.Entry{
			Kills:    entry.Kills,
			Duds:     entry.Duds,
			FreedMem: entry.FreedMem,
			Speed:    entry.Speed,
			Time:     entry.Time,
			Duration: entry.Duration,
			Date:     entry.Date,
		}
	}
	return out
}

func toBoardEntries(entries []score.Entry) []BoardEntry {
	out := make([]BoardEntry, len(entries))
	for i, entry := range entries {
		out[i] = BoardEntry{
			Kills:    entry.Kills,
			Duds:     entry.Duds,
			FreedMem: entry.FreedMem,
			Speed:    entry.Speed,
			Time:     entry.Time,
			Duration: entry.Duration,
			Date:     entry.Date,
		}
	}
	return out
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

func toLoadResult(board outbound.ScoreBoard) LoadResult {
	ranked := toBoard(board)
	return LoadResult{Entries: toBoardEntries(ranked.Scores), HighScore: ranked.HighScore()}
}

func toRecordResult(board *score.Board, kills int) RecordResult {
	return RecordResult{Entries: toBoardEntries(board.Scores), NewHighScore: board.IsNewHighScore(kills)}
}

func toScoreBoard(entries []BoardEntry) outbound.ScoreBoard {
	return outbound.ScoreBoard{Scores: toScoreEntries(entries)}
}

func toScoreSummary(request ReportRequest) outbound.ScoreSummary {
	return outbound.ScoreSummary{
		Kills:      request.Kills,
		Duds:       request.Duds,
		FreedMem:   request.FreedMem,
		Duration:   request.Duration,
		IsTopScore: request.NewHighScore,
		Entries:    toScoreEntries(request.Entries),
	}
}

func toScoreEntries(entries []BoardEntry) []outbound.ScoreEntry {
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
