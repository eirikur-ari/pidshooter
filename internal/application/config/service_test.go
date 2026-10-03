package config_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

// --- Service.Load ---

func TestLoadUsesFileDefaultsWhenNoOverrides(t *testing.T) {
	store := &testutil.FakeConfigStore{Config: outbound.Config{
		Process: outbound.ProcessConfig{IncludeRoot: testutil.Pointer(true)},
		Game:    outbound.GameConfig{ConfirmMode: testutil.Pointer(true), Speed: testutil.Pointer(3.0), TimeLimit: testutil.Pointer(45)},
	}}
	svc := config.NewService(store, config.Options{})

	cfg, err := svc.Load()

	require.NoError(t, err)
	assert.Equal(t, config.Result{Mode: outbound.ModeGame, Process: config.ProcessResult{IncludeRoot: true}, Game: config.GameResult{ConfirmMode: true, Speed: 3.0, TimeLimit: 45}}, cfg)
}

func TestLoadUsesHardcodedDefaultsWhenConfigFileNotFound(t *testing.T) {
	store := &testutil.FakeConfigStore{LoadErr: outbound.NotFoundError{}}
	svc := config.NewService(store, config.Options{})

	cfg, err := svc.Load()

	require.NoError(t, err)
	assert.Equal(t, config.Result{Mode: outbound.ModeGame, Process: config.ProcessResult{IncludeRoot: false}, Game: config.GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}, cfg)
}

func TestLoadReturnsWarningAndHardcodedDefaultsOnOtherLoadError(t *testing.T) {
	cause := errors.New("disk error")
	store := &testutil.FakeConfigStore{LoadErr: cause}
	svc := config.NewService(store, config.Options{})

	cfg, err := svc.Load()

	assert.Equal(t, config.Result{Mode: outbound.ModeGame, Process: config.ProcessResult{IncludeRoot: false}, Game: config.GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30}}, cfg)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreLoadFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorIs(t, err, cause)
}

func TestLoadOverridesWinOverFileDefaults(t *testing.T) {
	store := &testutil.FakeConfigStore{Config: outbound.Config{
		Process: outbound.ProcessConfig{IncludeRoot: testutil.Pointer(false)},
		Game:    outbound.GameConfig{ConfirmMode: testutil.Pointer(false), Speed: testutil.Pointer(3.0), TimeLimit: testutil.Pointer(45)},
	}}
	svc := config.NewService(store, config.Options{
		Game: config.GameOptions{
			ConfirmMode: testutil.Pointer(true),
			Speed:       testutil.Pointer(5.0),
			TimeLimit:   testutil.Pointer(60),
		},
		Process: config.ProcessOptions{IncludeRoot: testutil.Pointer(true)},
	})

	cfg, err := svc.Load()

	require.NoError(t, err)
	assert.Equal(t, config.Result{Mode: outbound.ModeGame, Process: config.ProcessResult{IncludeRoot: true}, Game: config.GameResult{ConfirmMode: true, Speed: 5.0, TimeLimit: 60}}, cfg)
}

func TestLoadOverridesWinOverHardcodedDefaultsWhenFileNotFound(t *testing.T) {
	store := &testutil.FakeConfigStore{LoadErr: outbound.NotFoundError{}}
	svc := config.NewService(store, config.Options{Game: config.GameOptions{Speed: testutil.Pointer(4.5)}})

	cfg, err := svc.Load()

	require.NoError(t, err)
	assert.Equal(t, config.Result{Mode: outbound.ModeGame, Process: config.ProcessResult{IncludeRoot: false}, Game: config.GameResult{ConfirmMode: false, Speed: 4.5, TimeLimit: 30}}, cfg)
}

func TestLoadPartialOverridesLeaveOtherFileDefaultsIntact(t *testing.T) {
	store := &testutil.FakeConfigStore{Config: outbound.Config{
		Process: outbound.ProcessConfig{IncludeRoot: testutil.Pointer(true)},
		Game:    outbound.GameConfig{ConfirmMode: testutil.Pointer(true), Speed: testutil.Pointer(3.0), TimeLimit: testutil.Pointer(45)},
	}}
	svc := config.NewService(store, config.Options{Game: config.GameOptions{Speed: testutil.Pointer(1.5)}})

	cfg, err := svc.Load()

	require.NoError(t, err)
	assert.Equal(t, config.Result{Mode: outbound.ModeGame, Process: config.ProcessResult{IncludeRoot: true}, Game: config.GameResult{ConfirmMode: true, Speed: 1.5, TimeLimit: 45}}, cfg)
}

func TestLoadReturnsInvalidRequestWithoutConsultingConfigFile(t *testing.T) {
	store := &testutil.FakeConfigStore{Config: outbound.Config{
		Process: outbound.ProcessConfig{IncludeRoot: testutil.Pointer(true)},
		Game:    outbound.GameConfig{ConfirmMode: testutil.Pointer(true), Speed: testutil.Pointer(3.0), TimeLimit: testutil.Pointer(45)},
	}}
	svc := config.NewService(store, config.Options{Game: config.GameOptions{Speed: testutil.Pointer(99.0)}})

	cfg, err := svc.Load()

	assert.Equal(t, config.Result{}, cfg, "an invalid request must short-circuit before any file defaults are merged in")
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}

func TestLoadReturnsInvalidRequestTimeLimitWithoutConsultingConfigFile(t *testing.T) {
	store := &testutil.FakeConfigStore{Config: outbound.Config{
		Game: outbound.GameConfig{TimeLimit: testutil.Pointer(45)},
	}}
	svc := config.NewService(store, config.Options{Game: config.GameOptions{TimeLimit: testutil.Pointer(-1)}})

	cfg, err := svc.Load()

	assert.Equal(t, config.Result{}, cfg, "an invalid request must short-circuit before any file defaults are merged in")
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
}

func TestLoadClassifiesInvalidFileValueAsWarning(t *testing.T) {
	store := &testutil.FakeConfigStore{Config: outbound.Config{
		Game: outbound.GameConfig{Speed: testutil.Pointer(99.0)},
	}}
	svc := config.NewService(store, config.Options{})

	cfg, err := svc.Load()

	assert.Equal(t, 2.0, cfg.Game.Speed, "an out-of-range file speed must not override the hardcoded default")
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorContains(t, err, "speed")
}
