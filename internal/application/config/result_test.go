package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil/helper"
)

// --- Result.apply ---

func TestResultApplyMapsAllPresentStoredFields(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.Config{
		Process: outbound.ProcessConfig{IncludeRoot: helper.Pointer(true)},
		Game:    outbound.GameConfig{ConfirmMode: helper.Pointer(true), Speed: helper.Pointer(2.5), TimeLimit: helper.Pointer(60)},
	}

	result, err := cfg.apply(stored, Options{})

	require.NoError(t, err)
	assert.Equal(t, Result{Process: ProcessResult{IncludeRoot: true}, Game: GameResult{ConfirmMode: true, Speed: 2.5, TimeLimit: 60}}, result)
}

func TestResultApplyEmptyInputLeavesCfgUnchanged(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}

	result, err := cfg.apply(outbound.Config{}, Options{})

	require.NoError(t, err)
	assert.Equal(t, cfg, result)
}

func TestResultApplyOnlySpeedSetLeavesTimeLimitAtDefault(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.Config{Game: outbound.GameConfig{Speed: helper.Pointer(3.0)}}

	result, err := cfg.apply(stored, Options{})

	require.NoError(t, err)
	assert.Equal(t, 3.0, result.Game.Speed)
	assert.Equal(t, 30, result.Game.TimeLimit, "TimeLimit must keep its hardcoded default when the file doesn't set it")
}

func TestResultApplyRejectsOutOfRangeSpeedAndKeepsDefault(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.Config{Game: outbound.GameConfig{Speed: helper.Pointer(99.0), TimeLimit: helper.Pointer(60)}}

	result, err := cfg.apply(stored, Options{})

	assert.Equal(t, 2.0, result.Game.Speed, "an out-of-range file speed must not override the hardcoded default")
	assert.Equal(t, 60, result.Game.TimeLimit, "a valid field must still apply even when a sibling field is rejected")
	assert.ErrorContains(t, err, "speed")
}

func TestResultApplyRejectsOutOfRangeTimeLimitAndKeepsDefault(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.Config{Game: outbound.GameConfig{TimeLimit: helper.Pointer(-1)}}

	result, err := cfg.apply(stored, Options{})

	assert.Equal(t, 30, result.Game.TimeLimit, "an out-of-range file time limit must not override the hardcoded default")
	assert.ErrorContains(t, err, "time_limit")
}

func TestResultApplyOverlaysValidStoredMode(t *testing.T) {
	cfg := Result{Mode: outbound.ModeGame, Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.Config{Mode: outbound.ModeLucky}

	result, err := cfg.apply(stored, Options{})

	require.NoError(t, err)
	assert.Equal(t, outbound.ModeLucky, result.Mode)
}

func TestResultApplyRejectsUnrecognizedStoredModeAndKeepsDefault(t *testing.T) {
	cfg := Result{Mode: outbound.ModeGame, Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.Config{Mode: outbound.Mode("wobble")}

	result, err := cfg.apply(stored, Options{})

	assert.Equal(t, outbound.ModeGame, result.Mode, "an unrecognized stored mode must not override the default")
	assert.ErrorContains(t, err, "mode")
}

func TestResultApplySetsAllowRootFromOptionsOnly(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false, AllowRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	opts := Options{Process: ProcessOptions{AllowRoot: helper.Pointer(true)}}

	result, err := cfg.apply(outbound.Config{}, opts)

	require.NoError(t, err)
	assert.True(t, result.Process.AllowRoot, "AllowRoot must be settable from options even though it is never read from the config file")
}

func TestResultApplyOptionsWinOverStored(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.Config{
		Process: outbound.ProcessConfig{IncludeRoot: helper.Pointer(false)},
		Game:    outbound.GameConfig{ConfirmMode: helper.Pointer(false), Speed: helper.Pointer(3.0), TimeLimit: helper.Pointer(45)},
	}
	opts := Options{
		Game: GameOptions{
			ConfirmMode: helper.Pointer(true),
			Speed:       helper.Pointer(5.0),
			TimeLimit:   helper.Pointer(60),
		},
		Process: ProcessOptions{IncludeRoot: helper.Pointer(true)},
	}

	result, err := cfg.apply(stored, opts)

	require.NoError(t, err)
	assert.Equal(t, Result{Process: ProcessResult{IncludeRoot: true}, Game: GameResult{ConfirmMode: true, Speed: 5.0, TimeLimit: 60}}, result)
}
