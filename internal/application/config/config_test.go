package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
)

func TestValidateNoPatterns(t *testing.T) {
	err := Config{Speed: 2.0, TimeLimit: 30}.Validate()
	assert.Error(t, err)
}

func TestValidateSpeedTooLow(t *testing.T) {
	err := Config{Patterns: []string{"proc"}, Speed: movement.MinSpeed - 0.1, TimeLimit: 30}.Validate()
	assert.Error(t, err)
}

func TestValidateSpeedTooHigh(t *testing.T) {
	err := Config{Patterns: []string{"proc"}, Speed: movement.MaxSpeed + 0.1, TimeLimit: 30}.Validate()
	assert.Error(t, err)
}

func TestValidateSpeedMinBoundary(t *testing.T) {
	err := Config{Patterns: []string{"proc"}, Speed: movement.MinSpeed, TimeLimit: 30}.Validate()
	assert.NoError(t, err)
}

func TestValidateSpeedMaxBoundary(t *testing.T) {
	err := Config{Patterns: []string{"proc"}, Speed: movement.MaxSpeed, TimeLimit: 30}.Validate()
	assert.NoError(t, err)
}

func TestValidateNegativeTimeLimit(t *testing.T) {
	err := Config{Patterns: []string{"proc"}, Speed: 2.0, TimeLimit: -1}.Validate()
	assert.Error(t, err)
}

func TestValidateZeroTimeLimitIsUnlimited(t *testing.T) {
	err := Config{Patterns: []string{"proc"}, Speed: 2.0, TimeLimit: 0}.Validate()
	assert.NoError(t, err)
}

func TestValidateValid(t *testing.T) {
	cfg := Config{Patterns: []string{"proc"}, Speed: 2.0, TimeLimit: 30}
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, []string{"proc"}, cfg.Patterns)
}