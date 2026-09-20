package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
)

func TestNewConfigNoPatterns(t *testing.T) {
	_, err := NewConfig(nil, false, 2.0, 30)
	assert.Error(t, err)
}

func TestNewConfigSpeedTooLow(t *testing.T) {
	_, err := NewConfig([]string{"proc"}, false, movement.MinSpeed-0.1, 30)
	assert.Error(t, err)
}

func TestNewConfigSpeedTooHigh(t *testing.T) {
	_, err := NewConfig([]string{"proc"}, false, movement.MaxSpeed+0.1, 30)
	assert.Error(t, err)
}

func TestNewConfigSpeedMinBoundary(t *testing.T) {
	_, err := NewConfig([]string{"proc"}, false, movement.MinSpeed, 30)
	assert.NoError(t, err)
}

func TestNewConfigSpeedMaxBoundary(t *testing.T) {
	_, err := NewConfig([]string{"proc"}, false, movement.MaxSpeed, 30)
	assert.NoError(t, err)
}

func TestNewConfigNegativeTimeLimit(t *testing.T) {
	_, err := NewConfig([]string{"proc"}, false, 2.0, -1)
	assert.Error(t, err)
}

func TestNewConfigZeroTimeLimitIsUnlimited(t *testing.T) {
	_, err := NewConfig([]string{"proc"}, false, 2.0, 0)
	assert.NoError(t, err)
}

func TestNewConfigValid(t *testing.T) {
	cfg, err := NewConfig([]string{"proc"}, false, 2.0, 30)
	assert.NoError(t, err)
	assert.Equal(t, []string{"proc"}, cfg.Patterns)
}