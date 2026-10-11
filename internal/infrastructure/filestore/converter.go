package filestore

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

func toScoreBoard(content scoreContent) outbound.ScoreBoard {
	entries := make([]outbound.ScoreEntry, len(content.Scores))
	for i, entry := range content.Scores {
		entries[i] = outbound.ScoreEntry{
			Kills:    entry.Kills,
			Duds:     entry.Duds,
			FreedMem: entry.FreedMem,
			Speed:    entry.Speed,
			Time:     entry.Time,
			Duration: entry.Duration,
			Date:     entry.Date,
		}
	}
	return outbound.ScoreBoard{Scores: entries}
}

func toScoreContent(board outbound.ScoreBoard) scoreContent {
	entries := make([]scoreEntry, len(board.Scores))
	for i, entry := range board.Scores {
		entries[i] = scoreEntry{
			Kills:    entry.Kills,
			Duds:     entry.Duds,
			FreedMem: entry.FreedMem,
			Speed:    entry.Speed,
			Time:     entry.Time,
			Duration: entry.Duration,
			Date:     entry.Date,
		}
	}
	return scoreContent{Version: currentScoreSchemaVersion, Scores: entries}
}

func toConfig(content configContent) outbound.Config {
	return outbound.Config{
		Mode: outbound.Mode(content.Mode),
		Process: outbound.ProcessConfig{
			IncludeRoot: util.ClonePointer(content.Process.IncludeRoot),
		},
		Game: outbound.GameConfig{
			ConfirmMode: util.ClonePointer(content.Game.ConfirmMode),
			Speed:       util.ClonePointer(content.Game.Speed),
			TimeLimit:   util.ClonePointer(content.Game.TimeLimit),
		},
	}
}

func toConfigContent(config outbound.Config) configContent {
	return configContent{
		Version: currentConfigSchemaVersion,
		Mode:    string(config.Mode),
		Process: processEntry{
			IncludeRoot: util.ClonePointer(config.Process.IncludeRoot),
		},
		Game: gameEntry{
			ConfirmMode: util.ClonePointer(config.Game.ConfirmMode),
			Speed:       util.ClonePointer(config.Game.Speed),
			TimeLimit:   util.ClonePointer(config.Game.TimeLimit),
		},
	}
}
