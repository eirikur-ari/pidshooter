package composition

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestNewRunner_CreatesErrorHandler(t *testing.T) {
	// When
	runner := NewRunner()

	// Then
	assert.NotNil(t, runner.errHandler)
}

func TestRunner_HandleError_AbsorbsWarningAndReturnsFatal(t *testing.T) {
	// Given
	runner := newRunnerFixture()
	warning := apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, "could not kill target", errors.New("boom"))
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))

	// When
	warningResult := runner.HandleError(warning)
	fatalResult := runner.HandleError(fatal)

	// Then
	assert.NoError(t, warningResult)
	assert.Same(t, fatal, fatalResult)
}

func TestRunner_buildPlay_ReturnsRunnerWhenAllDependenciesAvailable(t *testing.T) {
	// Given
	useTempHome(t)
	t.Setenv("TERM", "xterm-256color")

	// When
	runner, err := newRunnerFixture().buildPlay()

	// Then
	require.NoError(t, err)
	assert.NotNil(t, runner)
}

func TestRunner_buildPlay_ReturnsErrorWhenProcessManagerFails(t *testing.T) {
	// Given
	t.Setenv("PATH", "")

	// When
	runner, err := newRunnerFixture().buildPlay()

	// Then
	require.Error(t, err)
	assert.Nil(t, runner)
}

func TestRunner_buildPlay_ReturnsErrorWhenFileStoreFails(t *testing.T) {
	// Given
	testutil.UnsetEnv(t, "HOME")
	testutil.UnsetEnv(t, "XDG_CONFIG_HOME")

	// When
	runner, err := newRunnerFixture().buildPlay()

	// Then
	require.Error(t, err)
	assert.Nil(t, runner)
}

func TestRunner_buildPlay_ReturnsErrorWhenScreenFails(t *testing.T) {
	// Given
	useTempHome(t)
	testutil.UnsetEnv(t, "TERM")

	// When
	runner, err := newRunnerFixture().buildPlay()

	// Then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create screen")
	assert.Nil(t, runner)
}

func useTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	testutil.UnsetEnv(t, "XDG_CONFIG_HOME")
}
