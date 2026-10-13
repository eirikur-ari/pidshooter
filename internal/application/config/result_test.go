package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestNewResult_ReturnsBaselineDefaults(t *testing.T) {
	// When
	result := newResult()

	// Then
	assert.Equal(t, Result{
		Mode:    outbound.ModeGame,
		Process: ProcessResult{IncludeRoot: false, AllowRoot: false},
		Game:    GameResult{ConfirmMode: false, Speed: 2.0, TimeLimit: 30},
	}, result)
}

func TestResult_apply_OverlaysAllPresentStoredFields(t *testing.T) {
	// Given
	stored := newStoredConfigFixture()
	stored.Mode = outbound.ModeLucky

	expected := newResult()
	expected.Mode = outbound.ModeLucky
	expected.Process.IncludeRoot = true
	expected.Game = GameResult{ConfirmMode: true, Speed: 3.0, TimeLimit: 45}

	// When
	actual, err := newResult().apply(stored, Options{})

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestResult_apply_LeavesResultUnchangedWhenNothingIsProvided(t *testing.T) {
	// When
	actual, err := newResult().apply(outbound.Config{}, Options{})

	// Then
	require.NoError(t, err)
	assert.Equal(t, newResult(), actual)
}

func TestResult_apply_KeepsExistingValuesForFieldsNotStored(t *testing.T) {
	// Given
	stored := outbound.Config{Game: outbound.GameConfig{Speed: new(3.0)}}
	expected := newResult()
	expected.Game.Speed = 3.0

	// When
	actual, err := newResult().apply(stored, Options{})

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestResult_apply_SkipsInvalidStoredFields(t *testing.T) {
	tests := newApplySkippedTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual, _ := newResult().apply(test.stored, Options{})

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestResult_apply_ReportsInvalidStoredFields(t *testing.T) {
	tests := newApplyReportedTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			_, err := newResult().apply(test.stored, Options{})

			// Then
			require.Error(t, err)
			for _, fieldName := range test.expected {
				assert.ErrorContains(t, err, fieldName)
			}
		})
	}
}

func TestResult_apply_OptionsOverrideStoredConfig(t *testing.T) {
	// Given
	stored := newStoredConfigFixture()
	stored.Process.IncludeRoot = new(false)
	stored.Game.ConfirmMode = new(false)

	expected := newResult()
	expected.Process.IncludeRoot = true
	expected.Game = GameResult{ConfirmMode: true, Speed: 5.0, TimeLimit: 60}

	// When
	actual, err := newResult().apply(stored, newOptionsFixture())

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestResult_apply_SetsAllowRootFromOptionsOnly(t *testing.T) {
	// Given
	options := Options{Process: ProcessOptions{AllowRoot: new(true)}}

	// When
	actual, err := newResult().apply(outbound.Config{}, options)

	// Then
	require.NoError(t, err)
	assert.True(t, actual.Process.AllowRoot)
}

func newApplySkippedTestCase() []struct {
	name     string
	stored   outbound.Config
	expected Result
} {
	unchanged := newResult()
	withValidTimeLimit := newResult()
	withValidTimeLimit.Game.TimeLimit = 60

	tests := []struct {
		name     string
		stored   outbound.Config
		expected Result
	}{
		{
			"out-of-range speed leaves a valid sibling applied",
			outbound.Config{Game: outbound.GameConfig{Speed: new(99.0), TimeLimit: new(60)}},
			withValidTimeLimit,
		},
		{
			"out-of-range time limit",
			outbound.Config{Game: outbound.GameConfig{TimeLimit: new(-1)}},
			unchanged,
		},
		{
			"unrecognized mode",
			outbound.Config{Mode: outbound.Mode("wobble")},
			unchanged,
		},
		{
			"every validated field invalid",
			outbound.Config{
				Mode: outbound.Mode("wobble"),
				Game: outbound.GameConfig{Speed: new(99.0), TimeLimit: new(-1)},
			},
			unchanged,
		},
	}
	return tests
}

func newApplyReportedTestCase() []struct {
	name     string
	stored   outbound.Config
	expected []string
} {
	tests := []struct {
		name     string
		stored   outbound.Config
		expected []string
	}{
		{
			"out-of-range speed",
			outbound.Config{Game: outbound.GameConfig{Speed: new(99.0)}},
			[]string{speedFieldName},
		},
		{
			"out-of-range time limit",
			outbound.Config{Game: outbound.GameConfig{TimeLimit: new(-1)}},
			[]string{timeLimitFieldName},
		},
		{
			"unrecognized mode",
			outbound.Config{Mode: outbound.Mode("wobble")},
			[]string{modeFieldName},
		},
		{
			"every validated field invalid",
			outbound.Config{
				Mode: outbound.Mode("wobble"),
				Game: outbound.GameConfig{Speed: new(99.0), TimeLimit: new(-1)},
			},
			[]string{modeFieldName, speedFieldName, timeLimitFieldName},
		},
	}
	return tests
}
