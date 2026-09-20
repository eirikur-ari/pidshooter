package composition

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil/helper"
)

func TestCreateReturnsErrorWhenPSIsNotOnPath(t *testing.T) {
	t.Setenv("PATH", "")

	_, err := NewRunnerFactory().Create()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ps not found")
}

func TestCreateReturnsErrorWhenTerminalIsUnavailable(t *testing.T) {
	helper.UnsetEnv(t, "TERM")

	_, err := NewRunnerFactory().Create()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create screen")
}
