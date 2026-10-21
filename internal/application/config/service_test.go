package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestService_Load_AppliesOptionsOverStoredConfig(t *testing.T) {
	// Given
	store := configStoreMockSetup(newStoredConfigFixture(), nil)
	options := Options{Game: GameOptions{Speed: new(1.5)}}
	service := NewService(store)

	expected := newResult()
	expected.Process.IncludeRoot = true
	expected.Game = GameResult{ConfirmMode: true, Speed: 1.5, TimeLimit: 45}

	// When
	actual, err := service.Load(options)

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestService_Load_ReturnsDefaultsWhenStoreHasNoConfig(t *testing.T) {
	// Given
	store := configStoreMockSetup(outbound.Config{}, outbound.NotFoundError{})
	store.On("Save", mock.Anything).Return(nil)
	service := NewService(store)
	expected := newResult()

	// When
	actual, err := service.Load(Options{})

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestService_Load_AppliesConfigOptionsWhenStoreHasNoConfig(t *testing.T) {
	// Given
	store := configStoreMockSetup(outbound.Config{}, outbound.NotFoundError{})
	store.On("Save", mock.Anything).Return(nil)
	options := Options{Game: GameOptions{Speed: new(4.5)}}
	service := NewService(store)

	expected := newResult()
	expected.Game.Speed = 4.5

	// When
	actual, err := service.Load(options)

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestService_Load_CreatesConfigFileFromDefaultsWhenStoreHasNoConfig(t *testing.T) {
	// Given
	store := configStoreMockSetup(outbound.Config{}, outbound.NotFoundError{})
	defaults := newResult().toConfig()
	store.On("Save", defaults).Return(nil)
	service := NewService(store)

	// When
	_, err := service.Load(Options{Game: GameOptions{Speed: new(4.5)}})

	// Then
	require.NoError(t, err)
	store.AssertCalled(t, "Save", defaults)
}

func TestService_Load_ReturnsConfigAndErrorWhenCreatingConfigFileFails(t *testing.T) {
	tests := newCreatingConfigFileFailsTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			var appErr *apperror.Error
			cause := errors.New("read-only file system")
			store := configStoreMockSetup(outbound.Config{}, outbound.NotFoundError{})
			store.On("Save", mock.Anything).Return(cause)
			service := NewService(store)

			// When
			actual, err := service.Load(test.options)

			// Then
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
			assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
			assert.ErrorIs(t, err, cause)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestService_Load_ReturnsWarningAndAppliesOptionsWhenStoreLoadFails(t *testing.T) {
	// Given
	var appErr *apperror.Error
	cause := errors.New("disk error")
	store := configStoreMockSetup(outbound.Config{}, cause)
	options := Options{Game: GameOptions{Speed: new(4.5)}}
	service := NewService(store)

	expected := newResult()
	expected.Game.Speed = 4.5

	// When
	actual, err := service.Load(options)

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
		{"out-of-range speed", Options{Game: GameOptions{Speed: new(99.0)}}},
		{"out-of-range time limit", Options{Game: GameOptions{TimeLimit: new(-1)}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			var appErr *apperror.Error
			store := configStoreMockSetup(newStoredConfigFixture(), nil)
			service := NewService(store)

			// When
			actual, err := service.Load(test.options)

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
	store := configStoreMockSetup(outbound.Config{Game: outbound.GameConfig{Speed: new(99.0)}}, nil)
	service := NewService(store)

	// When
	actual, err := service.Load(Options{})

	// Then
	assert.Equal(t, newResult(), actual)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorContains(t, err, speedFieldName)
}

func newCreatingConfigFileFailsTestCase() []struct {
	name     string
	options  Options
	expected Result
} {
	withSpeed := newResult()
	withSpeed.Game.Speed = 4.5

	return []struct {
		name     string
		options  Options
		expected Result
	}{
		{"no options", Options{}, newResult()},
		{"options provided", Options{Game: GameOptions{Speed: new(4.5)}}, withSpeed},
	}
}
