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
	// regardless of which user is running pidshooter.
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

// defaultSpeed and defaultTimeLimit are the baseline gameplay parameters, used
// when neither the stored config nor the options override them.
const (
	defaultSpeed     = 2.0
	defaultTimeLimit = 30
)

// Names of the stored fields that are validated.
const (
	modeFieldName      = "mode"
	speedFieldName     = "speed"
	timeLimitFieldName = "time_limit"
)

// newResult returns a Result populated with the default gameplay and
// process-discovery parameters.
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

// apply overlays the stored config and then the options onto the Result; the
// options always win over the stored config. A stored field that fails
// validation is skipped, keeping the existing value. The returned error is
// nil, or an error naming every skipped stored field.
func (result Result) apply(stored outbound.Config, options Options) (Result, error) {
	result, err := result.fromStore(stored)
	return result.fromOptions(options), err
}

// fromStore overlays the present fields of the stored config onto the Result,
// skipping any field whose stored value fails validation. The returned error is
// nil, or an error naming every skipped field and why.
func (result Result) fromStore(stored outbound.Config) (Result, error) {
	storedGame := stored.Game
	rejected, err := validateStore(stored)

	if stored.Mode != "" && !slices.Contains(rejected, modeFieldName) {
		result.Mode = stored.Mode
	}
	if storedGame.ConfirmMode != nil {
		result.Game.ConfirmMode = *storedGame.ConfirmMode
	}
	if storedGame.Speed != nil && !slices.Contains(rejected, speedFieldName) {
		result.Game.Speed = *storedGame.Speed
	}
	if storedGame.TimeLimit != nil && !slices.Contains(rejected, timeLimitFieldName) {
		result.Game.TimeLimit = *storedGame.TimeLimit
	}
	if stored.Process.IncludeRoot != nil {
		result.Process.IncludeRoot = *stored.Process.IncludeRoot
	}

	return result, err
}

// fromOptions overlays the explicitly provided options onto the Result.
func (result Result) fromOptions(options Options) Result {
	if options.Game.ConfirmMode != nil {
		result.Game.ConfirmMode = *options.Game.ConfirmMode
	}
	if options.Game.Speed != nil {
		result.Game.Speed = *options.Game.Speed
	}
	if options.Game.TimeLimit != nil {
		result.Game.TimeLimit = *options.Game.TimeLimit
	}
	if options.Process.IncludeRoot != nil {
		result.Process.IncludeRoot = *options.Process.IncludeRoot
	}
	if options.Process.AllowRoot != nil {
		result.Process.AllowRoot = *options.Process.AllowRoot
	}
	return result
}

// validateStore returns the names of the stored fields that fail validation,
// along with an error naming each of them and why, or nil if all are valid.
func validateStore(stored outbound.Config) (rejected []string, err error) {
	var causes []error

	if stored.Mode != "" {
		switch stored.Mode {
		case outbound.ModeGame, outbound.ModeLucky, outbound.ModeList:
		default:
			rejected = append(rejected, modeFieldName)
			causes = append(causes, fmt.Errorf("mode must be one of %q, %q, %q, got: %q", outbound.ModeGame, outbound.ModeLucky, outbound.ModeList, stored.Mode))
		}
	}

	storedGame := stored.Game
	if storedGame.Speed != nil {
		if validationErr := movement.ValidateSpeed(*storedGame.Speed); validationErr != nil {
			rejected = append(rejected, speedFieldName)
			causes = append(causes, validationErr)
		}
	}
	if storedGame.TimeLimit != nil {
		if validationErr := game.ValidateTimeLimit(*storedGame.TimeLimit); validationErr != nil {
			rejected = append(rejected, timeLimitFieldName)
			causes = append(causes, validationErr)
		}
	}

	if len(rejected) == 0 {
		return nil, nil
	}
	return rejected, fmt.Errorf("stored config values ignored: %s: %w", strings.Join(rejected, ", "), errors.Join(causes...))
}
