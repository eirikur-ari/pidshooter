package tcellui_test

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
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

func TestRenderClearsStaleContentFromPreviousFrame(t *testing.T) {
	ui, screen := newTUI(t)

	ui.Render(outbound.FrameState{
		Targets: []outbound.TargetViewState{{X: 0, Y: 5, Tag: "[1234 victim]"}},
	})
	require.Contains(t, rowContent(screen, 5), "victim")

	ui.Render(outbound.FrameState{})

	assert.NotContains(t, rowContent(screen, 5), "victim",
		"a target drawn in a previous frame must not linger once it's no longer in the frame state")
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

// rowContent reads the visible characters on the given screen row.
func rowContent(screen tcell.SimulationScreen, row int) string {
	cells, w, _ := screen.GetContents()
	var sb strings.Builder
	for x := range w {
		if r := cells[row*w+x].Runes; len(r) > 0 {
			sb.WriteRune(r[0])
		}
	}
	return strings.TrimRight(sb.String(), " ")
}
