package filestore

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

func toScoreBoard(c scoreContent) outbound.ScoreBoard {
	entries := make([]outbound.ScoreEntry, len(c.Scores))
	for i, e := range c.Scores {
		entries[i] = outbound.ScoreEntry{
			Kills:    e.Kills,
			Duds:     e.Duds,
			FreedMem: e.FreedMem,
			Speed:    e.Speed,
			Time:     e.Time,
			Duration: e.Duration,
			Date:     e.Date,
		}
	}
	return outbound.ScoreBoard{Scores: entries}
}

func toScoreContent(sb outbound.ScoreBoard) scoreContent {
	entries := make([]scoreEntry, len(sb.Scores))
	for i, e := range sb.Scores {
		entries[i] = scoreEntry{
			Kills:    e.Kills,
			Duds:     e.Duds,
			FreedMem: e.FreedMem,
			Speed:    e.Speed,
			Time:     e.Time,
			Duration: e.Duration,
			Date:     e.Date,
		}
	}
	return scoreContent{Version: currentScoreSchemaVersion, Scores: entries}
}

func toConfigStoreResult(config configContent) outbound.ConfigStoreResult {
	return outbound.ConfigStoreResult{
		Mode: outbound.Mode(config.Mode),
		Process: outbound.ProcessConfig{
			IncludeRoot: util.ClonePtr(config.Process.IncludeRoot),
		},
		Game: outbound.GameConfig{
			ConfirmMode: util.ClonePtr(config.Game.ConfirmMode),
			Speed:       util.ClonePtr(config.Game.Speed),
			TimeLimit:   util.ClonePtr(config.Game.TimeLimit),
		},
	}
}

func toConfigContent(config outbound.ConfigStoreResult) configContent {
	return configContent{
		Version: currentConfigSchemaVersion,
		Mode:    string(config.Mode),
		Process: processEntry{
			IncludeRoot: util.ClonePtr(config.Process.IncludeRoot),
		},
		Game: configEntry{
			ConfirmMode: util.ClonePtr(config.Game.ConfirmMode),
			Speed:       util.ClonePtr(config.Game.Speed),
			TimeLimit:   util.ClonePtr(config.Game.TimeLimit),
		},
	}
}
