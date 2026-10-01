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
	// AllowRoot permits running pidshooter itself as root; never populated
	// from the stored config.
	AllowRoot bool
}

// Result holds the run parameters for a single invocation of pidshooter.
type Result struct {
	Mode    outbound.Mode
	Process ProcessResult
	Game    GameResult
}

// defaultSpeed and defaultTimeLimit are this application's baseline gameplay
// parameters, used when neither the stored config nor options override them.
const (
	defaultSpeed     = 2.0
	defaultTimeLimit = 30
)

// Stored field names shared between validateStore and fromStore.
const (
	fieldMode      = "mode"
	fieldSpeed     = "speed"
	fieldTimeLimit = "time_limit"
)

// newResult returns a Result populated with this application's default
// gameplay and process-discovery parameters.
func newResult() Result {
	return Result{
		Mode: outbound.ModeGame,
		Process: ProcessResult{
			IncludeRoot: false,
			AllowRoot:   false,
		},
		Game: GameResult{
			ConfirmMode: false,
			Speed:       defaultSpeed,
			TimeLimit:   defaultTimeLimit,
		},
	}
}

// apply overlays stored's and then opts's explicitly-provided fields onto
// cfg, in that priority order — opts always wins over stored. A stored
// field that fails domain validation is skipped, keeping cfg's existing
// value. The returned error is nil, or a plain error naming every
// skipped stored field.
func (cfg Result) apply(stored outbound.Config, opts Options) (Result, error) {
	cfg, err := cfg.fromStore(stored)
	return cfg.fromOptions(opts), err
}

// fromStore overlays stored's present fields onto cfg, skipping any field
// whose persisted value fails domain validation. The returned error is
// nil, or a plain error naming every skipped field and why.
func (cfg Result) fromStore(stored outbound.Config) (Result, error) {
	g := stored.Game
	rejected, err := validateStore(stored)

	if stored.Mode != "" && !slices.Contains(rejected, fieldMode) {
		cfg.Mode = stored.Mode
	}
	if g.ConfirmMode != nil {
		cfg.Game.ConfirmMode = *g.ConfirmMode
	}
	if g.Speed != nil && !slices.Contains(rejected, fieldSpeed) {
		cfg.Game.Speed = *g.Speed
	}
	if g.TimeLimit != nil && !slices.Contains(rejected, fieldTimeLimit) {
		cfg.Game.TimeLimit = *g.TimeLimit
	}
	if stored.Process.IncludeRoot != nil {
		cfg.Process.IncludeRoot = *stored.Process.IncludeRoot
	}

	return cfg, err
}

// fromOptions overlays opts's explicitly-provided fields onto cfg.
func (cfg Result) fromOptions(opts Options) Result {
	if opts.Game.ConfirmMode != nil {
		cfg.Game.ConfirmMode = *opts.Game.ConfirmMode
	}
	if opts.Game.Speed != nil {
		cfg.Game.Speed = *opts.Game.Speed
	}
	if opts.Game.TimeLimit != nil {
		cfg.Game.TimeLimit = *opts.Game.TimeLimit
	}
	if opts.Process.IncludeRoot != nil {
		cfg.Process.IncludeRoot = *opts.Process.IncludeRoot
	}
	if opts.Process.AllowRoot != nil {
		cfg.Process.AllowRoot = *opts.Process.AllowRoot
	}
	return cfg
}

// validateStore reports every present field in stored that fails domain
// validation. The returned error is nil, or a plain error naming every
// invalid field and why.
func validateStore(stored outbound.Config) (rejected []string, err error) {
	var causes []error

	if stored.Mode != "" {
		switch stored.Mode {
		case outbound.ModeGame, outbound.ModeLucky, outbound.ModeList:
		default:
			rejected = append(rejected, fieldMode)
			causes = append(causes, fmt.Errorf("mode must be one of %q, %q, %q, got: %q", outbound.ModeGame, outbound.ModeLucky, outbound.ModeList, stored.Mode))
		}
	}

	config := stored.Game
	if config.Speed != nil {
		if validationErr := movement.ValidateSpeed(*config.Speed); validationErr != nil {
			rejected = append(rejected, fieldSpeed)
			causes = append(causes, validationErr)
		}
	}
	if config.TimeLimit != nil {
		if validationErr := game.ValidateTimeLimit(*config.TimeLimit); validationErr != nil {
			rejected = append(rejected, fieldTimeLimit)
			causes = append(causes, validationErr)
		}
	}

	if len(rejected) == 0 {
		return nil, nil
	}
	return rejected, fmt.Errorf("stored config values ignored: %s: %w", strings.Join(rejected, ", "), errors.Join(causes...))
}
