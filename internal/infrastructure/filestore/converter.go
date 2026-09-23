package filestore

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
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

func toDefaultConfig(cd configContent) outbound.DefaultConfig {
	return outbound.DefaultConfig{
		Mode: outbound.Mode(cd.Mode),
		Game: outbound.GameConfig{
			ConfirmMode: cd.Game.ConfirmMode,
			Speed:       cd.Game.Speed,
			TimeLimit:   cd.Game.TimeLimit,
			IncludeRoot: cd.Game.IncludeRoot,
		},
	}
}

func toConfigContent(defaults outbound.DefaultConfig) configContent {
	return configContent{
		Version: currentConfigSchemaVersion,
		Mode:    string(defaults.Mode),
		Game: configEntry{
			ConfirmMode: defaults.Game.ConfirmMode,
			Speed:       defaults.Game.Speed,
			TimeLimit:   defaults.Game.TimeLimit,
			IncludeRoot: defaults.Game.IncludeRoot,
		},
	}
}
