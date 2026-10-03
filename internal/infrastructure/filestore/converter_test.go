package filestore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
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

// --- toConfig ---

func TestToConfigMapsFields(t *testing.T) {
	cd := configContent{
		Version: currentConfigSchemaVersion,
		Mode:    "lucky",
		Process: processEntry{
			IncludeRoot: testutil.Pointer(true),
		},
		Game: configEntry{
			ConfirmMode: testutil.Pointer(true),
			Speed:       testutil.Pointer(2.5),
			TimeLimit:   testutil.Pointer(60),
		},
	}

	defaults := toConfig(cd)

	assert.Equal(t, outbound.ModeLucky, defaults.Mode)
	assert.Equal(t, outbound.ProcessConfig{IncludeRoot: testutil.Pointer(true)}, defaults.Process)
	assert.Equal(t, outbound.GameConfig{ConfirmMode: testutil.Pointer(true), Speed: testutil.Pointer(2.5), TimeLimit: testutil.Pointer(60)}, defaults.Game)
}

func TestToConfigEmptyInput(t *testing.T) {
	defaults := toConfig(configContent{})

	assert.Empty(t, defaults)
}

// --- toConfigContent ---

func TestToConfigContentMapsFields(t *testing.T) {
	defaults := outbound.Config{
		Mode:    outbound.ModeLucky,
		Process: outbound.ProcessConfig{IncludeRoot: testutil.Pointer(true)},
		Game:    outbound.GameConfig{ConfirmMode: testutil.Pointer(true), Speed: testutil.Pointer(2.5), TimeLimit: testutil.Pointer(60)},
	}

	cd := toConfigContent(defaults)

	assert.Equal(t, currentConfigSchemaVersion, cd.Version)
	assert.Equal(t, "lucky", cd.Mode)
	assert.Equal(t, processEntry{IncludeRoot: testutil.Pointer(true)}, cd.Process)
	assert.Equal(t, configEntry{ConfirmMode: testutil.Pointer(true), Speed: testutil.Pointer(2.5), TimeLimit: testutil.Pointer(60)}, cd.Game)
}

func TestToConfigContentEmptyInput(t *testing.T) {
	cd := toConfigContent(outbound.Config{})

	assert.Equal(t, currentConfigSchemaVersion, cd.Version)
	assert.Equal(t, "", cd.Mode)
	assert.Empty(t, cd.Game)
}
