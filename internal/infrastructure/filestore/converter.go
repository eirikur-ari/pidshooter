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

func toConfig(config configContent) outbound.Config {
	return outbound.Config{
		Mode: outbound.Mode(config.Mode),
		Process: outbound.ProcessConfig{
			IncludeRoot: util.ClonePointer(config.Process.IncludeRoot),
		},
		Game: outbound.GameConfig{
			ConfirmMode: util.ClonePointer(config.Game.ConfirmMode),
			Speed:       util.ClonePointer(config.Game.Speed),
			TimeLimit:   util.ClonePointer(config.Game.TimeLimit),
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
		Game: configEntry{
			ConfirmMode: util.ClonePointer(config.Game.ConfirmMode),
			Speed:       util.ClonePointer(config.Game.Speed),
			TimeLimit:   util.ClonePointer(config.Game.TimeLimit),
		},
	}
}
