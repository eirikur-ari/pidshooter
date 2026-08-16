package service

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

func toProcessInfo(p outbound.ProcessInfo) process.Info {
	return process.NewInfo(p.Pid, p.Name, p.Rss)
}

func toProcessInfos(ps []outbound.ProcessInfo) []process.Info {
	out := make([]process.Info, len(ps))
	for i, p := range ps {
		out[i] = toProcessInfo(p)
	}
	return out
}

func toBoard(sb outbound.ScoreBoard) (*score.Board, *score.Tracker) {
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
