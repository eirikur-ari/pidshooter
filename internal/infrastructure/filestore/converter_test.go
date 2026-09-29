package filestore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil/helper"
)

// --- toScoreBoard ---

func TestToScoreBoardMapsFields(t *testing.T) {
	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	c := scoreContent{Version: currentScoreSchemaVersion, Scores: []scoreEntry{
		{Kills: 7, Duds: 2, FreedMem: 4096, Speed: 2.5, Time: 60, Duration: 45.0, Date: date},
	}}

	sb := toScoreBoard(c)

	require.Len(t, sb.Scores, 1)
	assert.Equal(t, outbound.ScoreEntry{Kills: 7, Duds: 2, FreedMem: 4096, Speed: 2.5, Time: 60, Duration: 45.0, Date: date}, sb.Scores[0])
}

func TestToScoreBoardEmptyInput(t *testing.T) {
	sb := toScoreBoard(scoreContent{})

	assert.Empty(t, sb.Scores)
}

// --- toScoreContent ---

func TestToScoreContentMapsFields(t *testing.T) {
	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	sb := outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 7, Duds: 2, FreedMem: 4096, Speed: 2.5, Time: 60, Duration: 45.0, Date: date},
	}}

	c := toScoreContent(sb)

	assert.Equal(t, currentScoreSchemaVersion, c.Version)
	require.Len(t, c.Scores, 1)
	assert.Equal(t, scoreEntry{Kills: 7, Duds: 2, FreedMem: 4096, Speed: 2.5, Time: 60, Duration: 45.0, Date: date}, c.Scores[0])
}

func TestToScoreContentEmptyInput(t *testing.T) {
	c := toScoreContent(outbound.ScoreBoard{})

	assert.Equal(t, currentScoreSchemaVersion, c.Version)
	assert.Empty(t, c.Scores)
}

// --- toConfigStoreResult ---

func TestToConfigStoreResultMapsFields(t *testing.T) {
	cd := configContent{
		Version: currentConfigSchemaVersion,
		Mode:    "yolo",
		Process: processEntry{
			IncludeRoot: helper.Ptr(true),
		},
		Game: configEntry{
			ConfirmMode: helper.Ptr(true),
			Speed:       helper.Ptr(2.5),
			TimeLimit:   helper.Ptr(60),
		},
	}

	defaults := toConfigStoreResult(cd)

	assert.Equal(t, outbound.ModeYolo, defaults.Mode)
	assert.Equal(t, outbound.ProcessConfig{IncludeRoot: helper.Ptr(true)}, defaults.Process)
	assert.Equal(t, outbound.GameConfig{ConfirmMode: helper.Ptr(true), Speed: helper.Ptr(2.5), TimeLimit: helper.Ptr(60)}, defaults.Game)
}

func TestToConfigStoreResultEmptyInput(t *testing.T) {
	defaults := toConfigStoreResult(configContent{})

	assert.Empty(t, defaults)
}

// --- toConfigContent ---

func TestToConfigContentMapsFields(t *testing.T) {
	defaults := outbound.ConfigStoreResult{
		Mode:    outbound.ModeYolo,
		Process: outbound.ProcessConfig{IncludeRoot: helper.Ptr(true)},
		Game:    outbound.GameConfig{ConfirmMode: helper.Ptr(true), Speed: helper.Ptr(2.5), TimeLimit: helper.Ptr(60)},
	}

	cd := toConfigContent(defaults)

	assert.Equal(t, currentConfigSchemaVersion, cd.Version)
	assert.Equal(t, "yolo", cd.Mode)
	assert.Equal(t, processEntry{IncludeRoot: helper.Ptr(true)}, cd.Process)
	assert.Equal(t, configEntry{ConfirmMode: helper.Ptr(true), Speed: helper.Ptr(2.5), TimeLimit: helper.Ptr(60)}, cd.Game)
}

func TestToConfigContentEmptyInput(t *testing.T) {
	cd := toConfigContent(outbound.ConfigStoreResult{})

	assert.Equal(t, currentConfigSchemaVersion, cd.Version)
	assert.Equal(t, "", cd.Mode)
	assert.Empty(t, cd.Game)
}
