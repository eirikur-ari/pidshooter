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

// newUI creates an initialised UI backed by a simulation screen and registers
// Cleanup to call ui.Cleanup when the test ends.
func newUI(t *testing.T) (*tcellui.UI, tcell.SimulationScreen) {
	t.Helper()
	screen := tcell.NewSimulationScreen("")
	screen.SetSize(80, 25)
	ui := tcellui.NewUI(screen)
	require.NoError(t, ui.Init())
	t.Cleanup(ui.Cleanup)
	return ui, screen
}

// nextEvent reads one event from the UI with a timeout so tests fail fast
// instead of blocking forever if the expected event is never produced.
func nextEvent(t *testing.T, ui *tcellui.UI) outbound.InputEvent {
	t.Helper()
	select {
	case ev := <-ui.Events():
		return ev
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for event from poll goroutine")
		return nil
	}
}

// rowContent reads the visible characters on the given screen row.
func rowContent(screen tcell.SimulationScreen, row int) string {
	cells, w, _ := screen.GetContents()
	var sb strings.Builder
	for x := 0; x < w; x++ {
		if r := cells[row*w+x].Runes; len(r) > 0 {
			sb.WriteRune(r[0])
		}
	}
	return strings.TrimRight(sb.String(), " ")
}

// TestPollGoroutineExitsAfterCleanup is a regression test for issue #6.
func TestPollGoroutineExitsAfterCleanup(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewUI(screen)
	before := runtime.NumGoroutine()

	require.NoError(t, ui.Init())

	for i := 0; i < 15; i++ {
		screen.InjectKey(tcell.KeyRune, 'a', tcell.ModNone)
	}
	time.Sleep(50 * time.Millisecond)

	ui.Cleanup()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Errorf("poll goroutine did not exit after Cleanup: want ≤%d goroutines, got %d",
		before, runtime.NumGoroutine())
}

// TestDrawHUD_NarrowTerminalSuppressesCenter is a regression test for issue #14.
func TestDrawHUDNarrowTerminalSuppressesCenter(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewUI(screen)
	require.NoError(t, ui.Init())
	defer ui.Cleanup()

	screen.SetSize(30, 25)
	ui.Render(outbound.FrameState{HUD: outbound.HUDState{}})

	cells, w, _ := screen.GetContents()
	var row0 strings.Builder
	for x := 0; x < w; x++ {
		if r := cells[x].Runes; len(r) > 0 {
			row0.WriteRune(r[0])
		}
	}
	got := row0.String()

	assert.Contains(t, got, "FREED")
	assert.Contains(t, got, "KILLS")
	assert.NotContains(t, got, "Highscore", "row 0 should NOT contain Highscore on narrow terminal")
}

// TestDrawHUD_WideTerminalDrawsAllThree verifies all three HUD elements are visible on wide terminals.
func TestDrawHUDWideTerminalDrawsAllThree(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewUI(screen)
	require.NoError(t, ui.Init())
	defer ui.Cleanup()

	ui.Render(outbound.FrameState{HUD: outbound.HUDState{}})

	cells, w, _ := screen.GetContents()
	var row0 strings.Builder
	for x := 0; x < w; x++ {
		if r := cells[x].Runes; len(r) > 0 {
			row0.WriteRune(r[0])
		}
	}
	got := row0.String()

	for _, want := range []string{"FREED", "Highscore", "KILLS"} {
		assert.Contains(t, got, want, "row 0 should contain %q on wide terminal", want)
	}
}

// TestRender_MultiByteLabel_ColumnLayout is a regression test for issue #13.
func TestRenderMultiByteLabelColumnLayout(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewUI(screen)
	require.NoError(t, ui.Init())
	defer ui.Cleanup()

	ui.Render(outbound.FrameState{
		Targets: []outbound.TargetViewState{
			{X: 0, Y: 2, Tag: "✦ KILLED ✦", Killing: true},
		},
	})

	cells, w, _ := screen.GetContents()
	assert.Equal(t, ' ', cells[2*w+1].Runes[0], "col 1 should be space (rune after ✦) — byte-offset bug in render loop?")
	assert.Equal(t, 'K', cells[2*w+2].Runes[0], "col 2 should be 'K' — byte-offset bug in render loop?")
}

// --- translateKey (exercised via poll) ---

func TestPollTranslatesEscape(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	ke, ok := nextEvent(t, ui).(outbound.KeyEvent)
	require.True(t, ok, "expected KeyEvent")
	assert.Equal(t, outbound.KeyEscape, ke.Key)
}

func TestPollTranslatesCtrlC(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyCtrlC, 0, tcell.ModNone)
	ke, ok := nextEvent(t, ui).(outbound.KeyEvent)
	require.True(t, ok, "expected KeyEvent")
	assert.Equal(t, outbound.KeyCtrlC, ke.Key)
}

func TestPollTranslatesCtrlZ(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyCtrlZ, 0, tcell.ModNone)
	ke, ok := nextEvent(t, ui).(outbound.KeyEvent)
	require.True(t, ok, "expected KeyEvent")
	assert.Equal(t, outbound.KeyCtrlZ, ke.Key)
}

func TestPollTranslatesRune(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	ke, ok := nextEvent(t, ui).(outbound.KeyEvent)
	require.True(t, ok, "expected KeyEvent")
	assert.Equal(t, outbound.KeyNone, ke.Key, "expected KeyNone for plain rune")
	assert.Equal(t, 'q', ke.Ch)
}

// --- poll: event-type routing ---

func TestPollMouseButton1EmitsClickEvent(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectMouse(5, 10, tcell.Button1, tcell.ModNone)
	ce, ok := nextEvent(t, ui).(outbound.ClickEvent)
	require.True(t, ok, "expected ClickEvent")
	assert.Equal(t, 5, ce.X)
	assert.Equal(t, 10, ce.Y)
}

func TestPollNonButton1DropsEvent(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectMouse(5, 10, tcell.Button2, tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'z', tcell.ModNone)
	ev := nextEvent(t, ui)
	_, isClick := ev.(outbound.ClickEvent)
	assert.False(t, isClick, "Button2 should not produce a ClickEvent")
	ke, ok := ev.(outbound.KeyEvent)
	require.True(t, ok, "expected KeyEvent after dropped Button2")
	assert.Equal(t, 'z', ke.Ch)
}

// --- drawStatusBar ---

func TestDrawStatusBarNormal(t *testing.T) {
	ui, screen := newUI(t)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{Alive: 3, Speed: 2.0},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	for _, want := range []string{"Targets:", "Speed:", "Click to kill"} {
		assert.Contains(t, got, want, "status bar should contain %q", want)
	}
}

func TestDrawStatusBarConfirming(t *testing.T) {
	ui, screen := newUI(t)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{
			Confirming: &outbound.ConfirmViewState{PID: 42, Name: "myapp"},
		},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	for _, want := range []string{"42", "myapp", "(Y)es", "(N)o"} {
		assert.Contains(t, got, want, "confirm bar should contain %q", want)
	}
}

func TestDrawStatusBarWithTimeLimit(t *testing.T) {
	ui, screen := newUI(t)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{Alive: 1, Speed: 1.0, TimeLimit: 30, TimeLeft: 15},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.Contains(t, got, "Time:")
	assert.Contains(t, got, "15s")
}

func TestDrawStatusBarNoTimeLimit(t *testing.T) {
	ui, screen := newUI(t)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{Alive: 1, Speed: 1.0, TimeLimit: 0},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.NotContains(t, got, "Time:")
}
