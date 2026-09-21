package config

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// Config holds run parameters, sourced from persisted defaults and
// overridden by explicit input.
type Config struct {
	Patterns    []string
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}

// NewConfig constructs a Config from the given parameters, returning an
// error if any parameter fails validation.
func NewConfig(patterns []string, confirmMode bool, speed float64, timeLimit int) (Config, error) {
	cfg := Config{Patterns: patterns, ConfirmMode: confirmMode, Speed: speed, TimeLimit: timeLimit}
	if err := cfg.validate(); err != nil {
		return Config{}, apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", err)
	}
	return cfg, nil
}

func (c Config) validate() error {
	if err := process.ValidatePatterns(c.Patterns); err != nil {
		return err
	}
	if err := movement.ValidateSpeed(c.Speed); err != nil {
		return err
	}
	return game.ValidateTimeLimit(c.TimeLimit)
}