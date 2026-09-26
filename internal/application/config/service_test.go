package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
	"github.com/eirikur-ari/pidshooter/internal/testutil/helper"
)

// --- Service.Load ---

func TestLoadUsesFileDefaultsWhenNoOverrides(t *testing.T) {
	store := &fake.ConfigStore{Result: outbound.ConfigStoreResult{
		Process: outbound.ProcessConfig{IncludeRoot: helper.Ptr(true)},
		Game:    outbound.GameConfig{ConfirmMode: helper.Ptr(true), Speed: helper.Ptr(3.0), TimeLimit: helper.Ptr(45)},
	}}
	svc := NewService(store)

	cfg, err := svc.Load(Request{})

	require.NoError(t, err)
	assert.Equal(t, Result{Process: ProcessResult{IncludeRoot: true}, Game: GameResult{ConfirmMode: true, Speed: 3.0, TimeLimit: 45}}, cfg)
}

func TestLoadUsesHardcodedDefaultsWhenConfigFileNotFound(t *testing.T) {
	store := &fake.ConfigStore{LoadErr: outbound.NotFoundError{}}
	svc := NewService(store)

	cfg, err := svc.Load(Request{})

	require.NoError(t, err)
	assert.Equal(t, Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}, cfg)
}

func TestLoadReturnsWarningAndHardcodedDefaultsOnOtherLoadError(t *testing.T) {
	cause := errors.New("disk error")
	store := &fake.ConfigStore{LoadErr: cause}
	svc := NewService(store)

	cfg, err := svc.Load(Request{})

	assert.Equal(t, Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}, cfg)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreLoadFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorIs(t, err, cause)
}

func TestLoadOverridesWinOverFileDefaults(t *testing.T) {
	store := &fake.ConfigStore{Result: outbound.ConfigStoreResult{
		Process: outbound.ProcessConfig{IncludeRoot: helper.Ptr(false)},
		Game:    outbound.GameConfig{ConfirmMode: helper.Ptr(false), Speed: helper.Ptr(3.0), TimeLimit: helper.Ptr(45)},
	}}
	svc := NewService(store)

	cfg, err := svc.Load(Request{
		Game: GameRequest{
			ConfirmMode: helper.Ptr(true),
			Speed:       helper.Ptr(5.0),
			TimeLimit:   helper.Ptr(60),
		},
		Process: ProcessRequest{IncludeRoot: helper.Ptr(true)},
	})

	require.NoError(t, err)
	assert.Equal(t, Result{Process: ProcessResult{IncludeRoot: true}, Game: GameResult{ConfirmMode: true, Speed: 5.0, TimeLimit: 60}}, cfg)
}

func TestLoadOverridesWinOverHardcodedDefaultsWhenFileNotFound(t *testing.T) {
	store := &fake.ConfigStore{LoadErr: outbound.NotFoundError{}}
	svc := NewService(store)

	cfg, err := svc.Load(Request{Game: GameRequest{Speed: helper.Ptr(4.5)}})

	require.NoError(t, err)
	assert.Equal(t, Result{Process: ProcessResult{IncludeRoot: false}, Game: GameResult{ConfirmMode: false, Speed: 4.5, TimeLimit: 30}}, cfg)
}

func TestLoadPartialOverridesLeaveOtherFileDefaultsIntact(t *testing.T) {
	store := &fake.ConfigStore{Result: outbound.ConfigStoreResult{
		Process: outbound.ProcessConfig{IncludeRoot: helper.Ptr(true)},
		Game:    outbound.GameConfig{ConfirmMode: helper.Ptr(true), Speed: helper.Ptr(3.0), TimeLimit: helper.Ptr(45)},
	}}
	svc := NewService(store)

	cfg, err := svc.Load(Request{Game: GameRequest{Speed: helper.Ptr(1.5)}})

	require.NoError(t, err)
	assert.Equal(t, Result{Process: ProcessResult{IncludeRoot: true}, Game: GameResult{ConfirmMode: true, Speed: 1.5, TimeLimit: 45}}, cfg)
}

func TestLoadReturnsInvalidRequestWithoutConsultingConfigFile(t *testing.T) {
	store := &fake.ConfigStore{Result: outbound.ConfigStoreResult{
		Process: outbound.ProcessConfig{IncludeRoot: helper.Ptr(true)},
		Game:    outbound.GameConfig{ConfirmMode: helper.Ptr(true), Speed: helper.Ptr(3.0), TimeLimit: helper.Ptr(45)},
	}}
	svc := NewService(store)

	cfg, err := svc.Load(Request{Game: GameRequest{Speed: helper.Ptr(99.0)}})

	assert.Equal(t, Result{}, cfg, "an invalid request must short-circuit before any file defaults are merged in")
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}

func TestLoadClassifiesInvalidFileValueAsWarning(t *testing.T) {
	store := &fake.ConfigStore{Result: outbound.ConfigStoreResult{
		Game: outbound.GameConfig{Speed: helper.Ptr(99.0)},
	}}
	svc := NewService(store)

	cfg, err := svc.Load(Request{})

	assert.Equal(t, 2.0, cfg.Game.Speed, "an out-of-range file speed must not override the hardcoded default")
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorContains(t, err, "speed")
}
