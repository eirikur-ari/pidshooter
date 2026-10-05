package composition

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestRunnerCreator_Create_ReturnsRunnerWhenAllDependenciesAvailable(t *testing.T) {
	useTempHome(t)
	t.Setenv("TERM", "xterm-256color")

	runner, err := NewRunnerCreator().Create([]string{"chrome"}, config.Options{})

	require.NoError(t, err)
	assert.NotNil(t, runner)
}

func TestRunnerCreator_Create_ReturnsErrorWhenProcessManagerFails(t *testing.T) {
	t.Setenv("PATH", "")

	runner, err := NewRunnerCreator().Create(nil, config.Options{})

	require.Error(t, err)
	assert.Nil(t, runner)
}

func TestRunnerCreator_Create_ReturnsErrorWhenFileStoreFails(t *testing.T) {
	testutil.UnsetEnv(t, "HOME")
	testutil.UnsetEnv(t, "XDG_CONFIG_HOME")

	runner, err := NewRunnerCreator().Create(nil, config.Options{})

	require.Error(t, err)
	assert.Nil(t, runner)
}

func TestRunnerCreator_Create_ReturnsErrorWhenScreenFails(t *testing.T) {
	useTempHome(t)
	testutil.UnsetEnv(t, "TERM")

	runner, err := NewRunnerCreator().Create(nil, config.Options{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create screen")
	assert.Nil(t, runner)
}

func useTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	testutil.UnsetEnv(t, "XDG_CONFIG_HOME")
}
