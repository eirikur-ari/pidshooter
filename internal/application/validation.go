package application

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
)

// validateConfig enforces domain-level constraints on a play configuration.
// Pattern validation is left to processSvc.FindProcesses, which already
// enforces it.
func validateConfig(cfg inbound.Config) error {
	if cfg.Speed < movement.MinSpeed || cfg.Speed > movement.MaxSpeed {
		return fmt.Errorf("speed must be between %g and %g, got: %g", movement.MinSpeed, movement.MaxSpeed, cfg.Speed)
	}
	if cfg.TimeLimit < 0 {
		return fmt.Errorf("time must be 0 or positive, got: %d", cfg.TimeLimit)
	}
	return nil
}