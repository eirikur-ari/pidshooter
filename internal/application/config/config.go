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
	// IncludeRoot additionally permits root-owned processes as targets,
	// regardless of the caller's own effective UID.
	IncludeRoot bool
}

// Validate returns an error if any field fails validation.
func (c Config) Validate() error {
	if err := c.validate(); err != nil {
		return apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", err)
	}
	return nil
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