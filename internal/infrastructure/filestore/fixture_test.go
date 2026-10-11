package filestore

import (
	"path/filepath"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func dateFixture() time.Time {
	return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
}

func newConfigFixture() outbound.Config {
	return outbound.Config{
		Mode:    outbound.ModeGame,
		Process: outbound.ProcessConfig{IncludeRoot: testutil.Pointer(true)},
		Game: outbound.GameConfig{
			ConfirmMode: testutil.Pointer(true),
			Speed:       testutil.Pointer(2.5),
			TimeLimit:   testutil.Pointer(60),
		},
	}
}

func newConfigContentFixture() configContent {
	return configContent{
		Version: currentConfigSchemaVersion,
		Mode:    "game",
		Process: processEntry{IncludeRoot: testutil.Pointer(true)},
		Game: gameEntry{
			ConfirmMode: testutil.Pointer(true),
			Speed:       testutil.Pointer(2.5),
			TimeLimit:   testutil.Pointer(60),
		},
	}
}

func newScoreBoardFixture() outbound.ScoreBoard {
	return outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 7, Duds: 2, FreedMem: 4096, Speed: 2.5, Time: 60, Duration: 45.0, Date: dateFixture()},
		{Kills: 3, Duds: 0, FreedMem: 1024, Speed: 1.0, Time: 30, Duration: 20.0, Date: dateFixture().Add(24 * time.Hour)},
	}}
}

func newScoreContentFixture() scoreContent {
	return scoreContent{Version: currentScoreSchemaVersion, Scores: []scoreEntry{
		{Kills: 7, Duds: 2, FreedMem: 4096, Speed: 2.5, Time: 60, Duration: 45.0, Date: dateFixture()},
		{Kills: 3, Duds: 0, FreedMem: 1024, Speed: 1.0, Time: 30, Duration: 20.0, Date: dateFixture().Add(24 * time.Hour)},
	}}
}

func newConfigFileFixture(dir string) *configFile {
	return newConfigFileAt(filepath.Join(dir, "config.yaml")).(*configFile)
}

func newScoreFileFixture(dir string) *scoreFile {
	return newScoreFileAt(filepath.Join(dir, "scores.json")).(*scoreFile)
}

func newFileFixture(dir string, maxSize int) file {
	return file{path: filepath.Join(dir, "data.txt"), maxSize: maxSize}
}
