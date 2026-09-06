package application

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
)

func TestValidateConfigNoPatterns(t *testing.T) {
	assert.Error(t, validateConfig(inbound.Config{Speed: 2.0, TimeLimit: 30}))
}

func TestValidateConfigSpeedTooLow(t *testing.T) {
	assert.Error(t, validateConfig(inbound.Config{Patterns: []string{"proc"}, Speed: movement.MinSpeed - 0.1}))
}

func TestValidateConfigSpeedTooHigh(t *testing.T) {
	assert.Error(t, validateConfig(inbound.Config{Patterns: []string{"proc"}, Speed: movement.MaxSpeed + 0.1}))
}

func TestValidateConfigSpeedMinBoundary(t *testing.T) {
	assert.NoError(t, validateConfig(inbound.Config{Patterns: []string{"proc"}, Speed: movement.MinSpeed}))
}

func TestValidateConfigSpeedMaxBoundary(t *testing.T) {
	assert.NoError(t, validateConfig(inbound.Config{Patterns: []string{"proc"}, Speed: movement.MaxSpeed}))
}

func TestValidateConfigNegativeTimeLimit(t *testing.T) {
	assert.Error(t, validateConfig(inbound.Config{Patterns: []string{"proc"}, Speed: 2.0, TimeLimit: -1}))
}

func TestValidateConfigZeroTimeLimitIsUnlimited(t *testing.T) {
	assert.NoError(t, validateConfig(inbound.Config{Patterns: []string{"proc"}, Speed: 2.0, TimeLimit: 0}))
}

func TestValidateConfigValid(t *testing.T) {
	assert.NoError(t, validateConfig(inbound.Config{Patterns: []string{"proc"}, Speed: 2.0, TimeLimit: 30}))
}
