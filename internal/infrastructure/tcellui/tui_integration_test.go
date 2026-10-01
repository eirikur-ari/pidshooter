//go:build integration

package tcellui_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
)

func TestIntegrationCleanupBeforeInitDoesNotPanicOnRealScreen(t *testing.T) {
	t.Setenv("TERM", "xterm")

	screen, err := tcell.NewScreen()
	require.NoError(t, err)
	ui := tcellui.NewTUI(screen)

	assert.NotPanics(t, ui.Cleanup)
}
