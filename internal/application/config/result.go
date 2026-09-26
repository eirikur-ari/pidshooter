package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
)

// GameResult holds game mode run parameters.
type GameResult struct {
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}

// ProcessResult holds process-discovery run parameters.
type ProcessResult struct {
	// IncludeRoot additionally permits root-owned processes as targets,
	// regardless of the caller's own effective UID.
	IncludeRoot bool
}

// Result holds the run parameters for a single invocation of pidshooter.
type Result struct {
	Process ProcessResult
	Game    GameResult
}

// newResult returns a Result populated with this domain's default gameplay
// and process-discovery parameters.
func newResult() Result {
	cfg := game.DefaultConfig()
	return Result{
		Process: ProcessResult{
			IncludeRoot: false,
		},
		Game: GameResult{
			ConfirmMode: cfg.Confirm,
			Speed:       cfg.Speed,
			TimeLimit:   cfg.TimeLimit,
		},
	}
}

// apply overlays stored's and then req's explicitly-provided fields onto
// cfg, in that priority order — req always wins over stored. A stored
// field that fails domain validation is skipped, keeping cfg's existing
// value. The returned error is nil, or a plain error naming every
// skipped stored field.
func (cfg Result) apply(stored outbound.ConfigStoreResult, req Request) (Result, error) {
	cfg, err := cfg.fromStore(stored)
	return cfg.fromRequest(req), err
}

// fromStore overlays stored's present fields onto cfg, skipping any field
// whose persisted value fails domain validation. The returned error is
// nil, or a plain error naming every skipped field and why.
func (cfg Result) fromStore(stored outbound.ConfigStoreResult) (Result, error) {
	g := stored.Game
	rejected, err := validateStore(g)

	if g.ConfirmMode != nil {
		cfg.Game.ConfirmMode = *g.ConfirmMode
	}
	if g.Speed != nil && !slices.Contains(rejected, "speed") {
		cfg.Game.Speed = *g.Speed
	}
	if g.TimeLimit != nil && !slices.Contains(rejected, "time_limit") {
		cfg.Game.TimeLimit = *g.TimeLimit
	}
	if stored.Process.IncludeRoot != nil {
		cfg.Process.IncludeRoot = *stored.Process.IncludeRoot
	}

	return cfg, err
}

// fromRequest overlays req's explicitly-provided fields onto cfg.
func (cfg Result) fromRequest(req Request) Result {
	if req.Game.ConfirmMode != nil {
		cfg.Game.ConfirmMode = *req.Game.ConfirmMode
	}
	if req.Game.Speed != nil {
		cfg.Game.Speed = *req.Game.Speed
	}
	if req.Game.TimeLimit != nil {
		cfg.Game.TimeLimit = *req.Game.TimeLimit
	}
	if req.Process.IncludeRoot != nil {
		cfg.Process.IncludeRoot = *req.Process.IncludeRoot
	}
	return cfg
}

// validateStore reports every present field in config that fails domain
// validation. The returned error is nil, or a plain error naming every
// invalid field and why.
func validateStore(config outbound.GameConfig) (rejected []string, err error) {
	var causes []error

	if config.Speed != nil {
		if validationErr := movement.ValidateSpeed(*config.Speed); validationErr != nil {
			rejected = append(rejected, "speed")
			causes = append(causes, validationErr)
		}
	}
	if config.TimeLimit != nil {
		if validationErr := game.ValidateTimeLimit(*config.TimeLimit); validationErr != nil {
			rejected = append(rejected, "time_limit")
			causes = append(causes, validationErr)
		}
	}

	if len(rejected) == 0 {
		return nil, nil
	}
	return rejected, fmt.Errorf("config file values ignored: %s: %w", strings.Join(rejected, ", "), errors.Join(causes...))
}
