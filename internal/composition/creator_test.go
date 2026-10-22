package composition

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestNewCreator_HoldsTheBuildVersionOfTheRunningProgramAndAnErrorHandler(t *testing.T) {
	// Given
	expected := newBuildInfo().Version()

	// When
	creator := NewCreator()

	// Then
	assert.Equal(t, expected, creator.buildVersion)
	assert.NotNil(t, creator.errHandler)
}

func TestCreator_Create_PlayReportsItsBuildErrorWhileListIsUnaffected(t *testing.T) {
	// Given
	t.Setenv("PATH", "")
	testutil.UnsetEnv(t, "TERM")
	testutil.UnsetEnv(t, "HOME")
	testutil.UnsetEnv(t, "XDG_CONFIG_HOME")

	// When
	program := newCreatorFixture().Create()
	listErr := program.Run([]string{"list", "proc"})
	playErr := program.Run([]string{"play", "proc"})

	// Then
	assert.EqualError(t, listErr, "unimplemented")
	require.Error(t, playErr)
	assert.NotEqual(t, "unimplemented", playErr.Error())
}

func TestCreator_Create_LogsAnErrorOfACommandOnce(t *testing.T) {
	// Given
	handler, logger := newHandlerFixture()
	creator := &Creator{errHandler: handler}

	// When
	err := creator.Create().Run([]string{"kill", "proc"})

	// Then
	assert.EqualError(t, err, "unimplemented")
	assert.Equal(t, []string{"unimplemented"}, logger.Errored)
}

func TestCreator_buildPlay_ReturnsRunnerWhenAllDependenciesAvailable(t *testing.T) {
	// Given
	useTempHome(t)
	t.Setenv("TERM", "xterm-256color")

	// When
	runner, err := newCreatorFixture().buildPlay()

	// Then
	require.NoError(t, err)
	assert.NotNil(t, runner)
}

func TestCreator_buildPlay_ReturnsErrorWhenProcessManagerFails(t *testing.T) {
	// Given
	t.Setenv("PATH", "")

	// When
	runner, err := newCreatorFixture().buildPlay()

	// Then
	require.Error(t, err)
	assert.Nil(t, runner)
}

func TestCreator_buildPlay_ReturnsErrorWhenFileStoreFails(t *testing.T) {
	// Given
	testutil.UnsetEnv(t, "HOME")
	testutil.UnsetEnv(t, "XDG_CONFIG_HOME")

	// When
	runner, err := newCreatorFixture().buildPlay()

	// Then
	require.Error(t, err)
	assert.Nil(t, runner)
}

func TestCreator_buildPlay_ReturnsErrorWhenScreenFails(t *testing.T) {
	// Given
	useTempHome(t)
	testutil.UnsetEnv(t, "TERM")

	// When
	runner, err := newCreatorFixture().buildPlay()

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
