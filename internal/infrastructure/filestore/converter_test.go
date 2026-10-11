package filestore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestToScoreBoard_MapsEveryField(t *testing.T) {
	// Given
	expected := newScoreBoardFixture()

	// When
	board := toScoreBoard(newScoreContentFixture())

	// Then
	assert.Equal(t, expected, board)
}

func TestToScoreContent_MapsEveryFieldAndSetsCurrentSchemaVersion(t *testing.T) {
	// Given
	expected := newScoreContentFixture()

	// When
	content := toScoreContent(newScoreBoardFixture())

	// Then
	assert.Equal(t, expected, content)
}

func TestToConfig_MapsEveryField(t *testing.T) {
	// Given
	expected := newConfigFixture()

	// When
	config := toConfig(newConfigContentFixture())

	// Then
	assert.Equal(t, expected, config)
}

func TestToConfig_ClonesPointerFields(t *testing.T) {
	// Given
	content := newConfigContentFixture()
	expected := newConfigContentFixture()

	// When
	config := toConfig(content)
	*config.Game.Speed = 9.9
	*config.Game.ConfirmMode = false
	*config.Game.TimeLimit = 1
	*config.Process.IncludeRoot = false

	// Then
	assert.Equal(t, expected, content)
}

func TestToConfigContent_MapsEveryFieldAndSetsCurrentSchemaVersion(t *testing.T) {
	// Given
	expected := newConfigContentFixture()

	// When
	content := toConfigContent(newConfigFixture())

	// Then
	assert.Equal(t, expected, content)
}

func TestToConfigContent_ClonesPointerFields(t *testing.T) {
	// Given
	config := newConfigFixture()
	expected := newConfigFixture()

	// When
	content := toConfigContent(config)
	*content.Game.Speed = 9.9
	*content.Game.ConfirmMode = false
	*content.Game.TimeLimit = 1
	*content.Process.IncludeRoot = false

	// Then
	assert.Equal(t, expected, config)
}

func TestToConfigContent_LeavesUnsetFieldsNil(t *testing.T) {
	// When
	content := toConfigContent(outbound.Config{})

	// Then
	assert.Equal(t, currentConfigSchemaVersion, content.Version)
	assert.Nil(t, content.Process.IncludeRoot)
	assert.Nil(t, content.Game.ConfirmMode)
	assert.Nil(t, content.Game.Speed)
	assert.Nil(t, content.Game.TimeLimit)
}
