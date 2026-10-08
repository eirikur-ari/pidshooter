package osprocess

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestPinReturnsHandleForExistingPID(t *testing.T) {
	h, err := newTestProcess(t).Pin(os.Getpid())
	require.NoError(t, err)
	assert.NotNil(t, h)
}

func TestReleaseSucceedsForExistingPID(t *testing.T) {
	h, err := newTestProcess(t).Pin(os.Getpid())
	require.NoError(t, err)

	assert.NoError(t, h.Release())
}

func TestPinReturnsNotFoundForNonexistentPID(t *testing.T) {
	h, err := newTestProcess(t).Pin(999999)

	assert.Nil(t, h)
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
}

func TestKillerNonexistentPID(t *testing.T) {
	proc, err := os.FindProcess(999999)
	require.NoError(t, err)

	err = (&processHandle{proc: proc}).Kill()
	require.Error(t, err)
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
}
