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

// Mode identifies which run mode a Result resolves to.
type Mode string

// Mode's possible values.
const (
	ModeGame Mode = "game"
	ModeYolo Mode = "yolo"
	ModeList Mode = "list"
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
	Mode    Mode
	Process ProcessResult
	Game    GameResult
}

// defaultSpeed and defaultTimeLimit are this application's baseline gameplay
// parameters, used when neither a config file nor a request overrides them.
const (
	defaultSpeed     = 2.0
	defaultTimeLimit = 30
)

// newResult returns a Result populated with this application's default
// gameplay and process-discovery parameters.
func newResult() Result {
	return Result{
		Mode: ModeGame,
		Process: ProcessResult{
			IncludeRoot: false,
		},
		Game: GameResult{
			ConfirmMode: false,
			Speed:       defaultSpeed,
			TimeLimit:   defaultTimeLimit,
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
	rejected, err := validateStore(stored)

	if stored.Mode != "" && !slices.Contains(rejected, "mode") {
		cfg.Mode = Mode(stored.Mode)
	}
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

// validateStore reports every present field in stored that fails domain
// validation. The returned error is nil, or a plain error naming every
// invalid field and why.
func validateStore(stored outbound.ConfigStoreResult) (rejected []string, err error) {
	var causes []error

	if stored.Mode != "" {
		switch stored.Mode {
		case outbound.ModeGame, outbound.ModeYolo, outbound.ModeList:
		default:
			rejected = append(rejected, "mode")
			causes = append(causes, fmt.Errorf("mode must be one of %q, %q, %q, got: %q", outbound.ModeGame, outbound.ModeYolo, outbound.ModeList, stored.Mode))
		}
	}

	config := stored.Game
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
