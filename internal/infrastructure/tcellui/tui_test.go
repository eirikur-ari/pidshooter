package tcellui_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
)

func TestNewTUIPanicsOnNilScreen(t *testing.T) {
	assert.Panics(t, func() { tcellui.NewTUI(nil) })
}

func TestCleanupBeforeInitDoesNotPanicOnRealScreen(t *testing.T) {
	screen, err := tcell.NewScreen()
	require.NoError(t, err)
	ui := tcellui.NewTUI(screen)

	assert.NotPanics(t, ui.Cleanup)
}

func TestInitSecondCallReturnsErrorAndDoesNotSpawnSecondPollGoroutine(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewTUI(screen)
	before := runtime.NumGoroutine()

	require.NoError(t, ui.Init())
	require.Error(t, ui.Init())

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, before+1, runtime.NumGoroutine(), "a second Init call must not spawn a second poll goroutine")

	ui.Cleanup()
}

func TestChromeSizeReservesOneRowTopAndBottom(t *testing.T) {
	ui, _ := newTUI(t)
	chrome := ui.ChromeSize()
	assert.Equal(t, 1, chrome.Top)
	assert.Equal(t, 1, chrome.Bottom)
}

func TestWindowSizeReturnsScreenDimensions(t *testing.T) {
	ui, screen := newTUI(t)
	screen.SetSize(42, 17)
	window := ui.WindowSize()
	assert.Equal(t, 42, window.Width)
	assert.Equal(t, 17, window.Height)
}

// newTUI creates an initialised TUI backed by a simulation screen and registers
// Cleanup to call ui.Cleanup when the test ends.
func newTUI(t *testing.T) (*tcellui.TUI, tcell.SimulationScreen) {
	t.Helper()
	screen := tcell.NewSimulationScreen("")
	screen.SetSize(80, 25)
	ui := tcellui.NewTUI(screen)
	require.NoError(t, ui.Init())
	t.Cleanup(ui.Cleanup)
	return ui, screen
}
