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
	stored := outbound.ConfigStoreResult{
		Process: outbound.ProcessConfig{IncludeRoot: helper.Ptr(true)},
		Game:    outbound.GameConfig{ConfirmMode: helper.Ptr(true), Speed: helper.Ptr(2.5), TimeLimit: helper.Ptr(60)},
	}

	result, err := cfg.apply(stored, Request{})

	require.NoError(t, err)
	assert.Equal(t, Result{Process: ProcessResult{IncludeRoot: true}, Game: GameResult{ConfirmMode: true, Speed: 2.5, TimeLimit: 60}}, result)
}

func TestResultApplyEmptyInputLeavesCfgUnchanged(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}

	result, err := cfg.apply(outbound.ConfigStoreResult{}, Request{})

	require.NoError(t, err)
	assert.Equal(t, cfg, result)
}

func TestResultApplyOnlySpeedSetLeavesTimeLimitAtDefault(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.ConfigStoreResult{Game: outbound.GameConfig{Speed: helper.Ptr(3.0)}}

	result, err := cfg.apply(stored, Request{})

	require.NoError(t, err)
	assert.Equal(t, 3.0, result.Game.Speed)
	assert.Equal(t, 30, result.Game.TimeLimit, "TimeLimit must keep its hardcoded default when the file doesn't set it")
}

func TestResultApplyRejectsOutOfRangeSpeedAndKeepsDefault(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.ConfigStoreResult{Game: outbound.GameConfig{Speed: helper.Ptr(99.0), TimeLimit: helper.Ptr(60)}}

	result, err := cfg.apply(stored, Request{})

	assert.Equal(t, 2.0, result.Game.Speed, "an out-of-range file speed must not override the hardcoded default")
	assert.Equal(t, 60, result.Game.TimeLimit, "a valid field must still apply even when a sibling field is rejected")
	assert.ErrorContains(t, err, "speed")
}

func TestResultApplyRejectsOutOfRangeTimeLimitAndKeepsDefault(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.ConfigStoreResult{Game: outbound.GameConfig{TimeLimit: helper.Ptr(-1)}}

	result, err := cfg.apply(stored, Request{})

	assert.Equal(t, 30, result.Game.TimeLimit, "an out-of-range file time limit must not override the hardcoded default")
	assert.ErrorContains(t, err, "time_limit")
}

func TestResultApplyOverlaysValidStoredMode(t *testing.T) {
	cfg := Result{Mode: ModeGame, Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.ConfigStoreResult{Mode: outbound.ModeYolo}

	result, err := cfg.apply(stored, Request{})

	require.NoError(t, err)
	assert.Equal(t, ModeYolo, result.Mode)
}

func TestResultApplyRejectsUnrecognizedStoredModeAndKeepsDefault(t *testing.T) {
	cfg := Result{Mode: ModeGame, Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.ConfigStoreResult{Mode: outbound.Mode("wobble")}

	result, err := cfg.apply(stored, Request{})

	assert.Equal(t, ModeGame, result.Mode, "an unrecognized stored mode must not override the default")
	assert.ErrorContains(t, err, "mode")
}

func TestResultApplySetsAllowRootFromRequestOnly(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false, AllowRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	req := Request{Process: ProcessRequest{AllowRoot: helper.Ptr(true)}}

	result, err := cfg.apply(outbound.ConfigStoreResult{}, req)

	require.NoError(t, err)
	assert.True(t, result.Process.AllowRoot, "AllowRoot must be settable from the request even though it is never read from the config file")
}

func TestResultApplyRequestWinsOverStored(t *testing.T) {
	cfg := Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}
	stored := outbound.ConfigStoreResult{
		Process: outbound.ProcessConfig{IncludeRoot: helper.Ptr(false)},
		Game:    outbound.GameConfig{ConfirmMode: helper.Ptr(false), Speed: helper.Ptr(3.0), TimeLimit: helper.Ptr(45)},
	}
	req := Request{
		Game: GameRequest{
			ConfirmMode: helper.Ptr(true),
			Speed:       helper.Ptr(5.0),
			TimeLimit:   helper.Ptr(60),
		},
		Process: ProcessRequest{IncludeRoot: helper.Ptr(true)},
	}

	result, err := cfg.apply(stored, req)

	require.NoError(t, err)
	assert.Equal(t, Result{Process: ProcessResult{IncludeRoot: true}, Game: GameResult{ConfirmMode: true, Speed: 5.0, TimeLimit: 60}}, result)
}
