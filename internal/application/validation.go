package application

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// validateConfig enforces domain-level constraints on a play configuration.
func validateConfig(cfg inbound.Config) error {
	if err := process.ValidatePatterns(cfg.Patterns); err != nil {
		return err
	}
	if err := movement.ValidateSpeed(cfg.Speed); err != nil {
		return err
	}
	return game.ValidateTimeLimit(cfg.TimeLimit)
}
