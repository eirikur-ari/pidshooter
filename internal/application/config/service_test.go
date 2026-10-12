package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestService_Load_AppliesOptionsOverStoredConfig(t *testing.T) {
	// Given
	store := configStoreMockSetup(newStoredConfigFixture(), nil)
	options := Options{Game: GameOptions{Speed: testutil.Pointer(1.5)}}
	service := NewService(store, options)

	expected := newResult()
	expected.Process.IncludeRoot = true
	expected.Game = GameResult{ConfirmMode: true, Speed: 1.5, TimeLimit: 45}

	// When
	actual, err := service.Load()

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestService_Load_ReturnsDefaultsWhenStoreHasNoConfig(t *testing.T) {
	// Given
	store := configStoreMockSetup(outbound.Config{}, outbound.NotFoundError{})
	service := NewService(store, Options{})
	expected := newResult()

	// When
	actual, err := service.Load()

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestService_Load_AppliesOptionsWhenStoreHasNoConfig(t *testing.T) {
	// Given
	store := configStoreMockSetup(outbound.Config{}, outbound.NotFoundError{})
	options := Options{Game: GameOptions{Speed: testutil.Pointer(4.5)}}
	service := NewService(store, options)

	expected := newResult()
	expected.Game.Speed = 4.5

	// When
	actual, err := service.Load()

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestService_Load_ReturnsWarningAndAppliesOptionsWhenStoreLoadFails(t *testing.T) {
	// Given
	var appErr *apperror.Error
	cause := errors.New("disk error")
	store := configStoreMockSetup(outbound.Config{}, cause)
	options := Options{Game: GameOptions{Speed: testutil.Pointer(4.5)}}
	service := NewService(store, options)

	expected := newResult()
	expected.Game.Speed = 4.5

	// When
	actual, err := service.Load()

	// Then
	assert.Equal(t, expected, actual)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorIs(t, err, cause)
}

func TestService_Load_ReturnsFatalErrorWithoutReadingStoreWhenOptionsAreInvalid(t *testing.T) {
	tests := []struct {
		name    string
		options Options
	}{
		{"out-of-range speed", Options{Game: GameOptions{Speed: testutil.Pointer(99.0)}}},
		{"out-of-range time limit", Options{Game: GameOptions{TimeLimit: testutil.Pointer(-1)}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			var appErr *apperror.Error
			store := configStoreMockSetup(newStoredConfigFixture(), nil)
			service := NewService(store, test.options)

			// When
			actual, err := service.Load()

			// Then
			assert.Equal(t, Result{}, actual)
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
			assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
			store.AssertNotCalled(t, "Load")
		})
	}
}

func TestService_Load_ReturnsWarningAndKeepsDefaultWhenStoredValueIsInvalid(t *testing.T) {
	// Given
	var appErr *apperror.Error
	store := configStoreMockSetup(outbound.Config{Game: outbound.GameConfig{Speed: testutil.Pointer(99.0)}}, nil)
	service := NewService(store, Options{})

	// When
	actual, err := service.Load()

	// Then
	assert.Equal(t, newResult(), actual)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorContains(t, err, speedFieldName)
}
