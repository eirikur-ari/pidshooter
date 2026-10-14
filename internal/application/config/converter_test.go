package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestResult_toConfig_ReturnsEveryStoredField(t *testing.T) {
	// Given
	result := newResult()
	result.Process.AllowRoot = true

	expected := outbound.Config{
		Mode:    outbound.ModeGame,
		Process: outbound.ProcessConfig{IncludeRoot: new(false)},
		Game: outbound.GameConfig{
			ConfirmMode: new(false),
			Speed:       new(defaultSpeed),
			TimeLimit:   new(defaultTimeLimit),
		},
	}

	// When
	actual := result.toConfig()

	// Then
	assert.Equal(t, expected, actual)
}
