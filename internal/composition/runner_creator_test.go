package composition

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestCreateReturnsErrorWhenPSIsNotOnPath(t *testing.T) {
	t.Setenv("PATH", "")

	_, err := NewRunnerCreator().Create(nil, config.Options{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ps not found")
}

func TestCreateReturnsErrorWhenTerminalIsUnavailable(t *testing.T) {
	testutil.UnsetEnv(t, "TERM")

	_, err := NewRunnerCreator().Create(nil, config.Options{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create screen")
}
