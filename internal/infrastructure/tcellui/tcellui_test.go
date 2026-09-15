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

func TestPollGoroutineExitsAfterCleanup(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewUI(screen)
	before := runtime.NumGoroutine()

	require.NoError(t, ui.Init())

	for range 15 {
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

// TestCleanupBeforeInitDoesNotPanicOnRealScreen is a regression test for
// docs/tcellui-findings.md Finding 4. It deliberately uses a real
// tcell.Screen rather than the simulation screen every other test in this
// file uses: SimulationScreen.Fini nil-guards its internal quit channel,
// which a real screen does not, so this bug is invisible to a suite that
// only ever exercises the simulation double.
func TestCleanupBeforeInitDoesNotPanicOnRealScreen(t *testing.T) {
	screen, err := tcell.NewScreen()
	require.NoError(t, err)
	ui := tcellui.NewUI(screen)

	assert.NotPanics(t, ui.Cleanup)
}

func TestDrawHUDNarrowTerminalSuppressesCenter(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewUI(screen)
	require.NoError(t, ui.Init())
	defer ui.Cleanup()

	screen.SetSize(30, 25)
	ui.Render(outbound.FrameState{HUD: outbound.HUDState{}})

	cells, w, _ := screen.GetContents()
	var row0 strings.Builder
	for x := range w {
		if r := cells[x].Runes; len(r) > 0 {
			row0.WriteRune(r[0])
		}
	}
	got := row0.String()

	assert.Contains(t, got, "FREED")
	assert.Contains(t, got, "KILLS")
	assert.NotContains(t, got, "Highscore", "row 0 should NOT contain Highscore on narrow terminal")
}

func TestDrawHUDWideTerminalDrawsAllThree(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewUI(screen)
	require.NoError(t, ui.Init())
	defer ui.Cleanup()

	ui.Render(outbound.FrameState{HUD: outbound.HUDState{}})

	cells, w, _ := screen.GetContents()
	var row0 strings.Builder
	for x := range w {
		if r := cells[x].Runes; len(r) > 0 {
			row0.WriteRune(r[0])
		}
	}
	got := row0.String()

	for _, want := range []string{"FREED", "Highscore", "KILLS"} {
		assert.Contains(t, got, want, "row 0 should contain %q on wide terminal", want)
	}
}

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

func TestRenderWideRuneTagDoesNotDropCharacters(t *testing.T) {
	ui, screen := newUI(t)
	ui.Render(outbound.FrameState{
		Targets: []outbound.TargetViewState{{X: 0, Y: 5, Tag: "[9 日本語]"}},
	})

	got := rowContent(screen, 5)
	for _, want := range []string{"[", "9", "日", "本", "語", "]"} {
		assert.Contains(t, got, want, "wide-rune tag should render %q without dropping characters", want)
	}
}

func TestRenderClipsTargetAtHUDRow(t *testing.T) {
	ui, screen := newUI(t)
	ui.Render(outbound.FrameState{
		Targets: []outbound.TargetViewState{{X: 2, Y: 0, Tag: "[1234 victim]"}},
		HUD:     outbound.HUDState{Kills: 3},
	})

	got := rowContent(screen, 0)
	assert.NotContains(t, got, "victim", "target tag must not be drawn on the HUD row")
}

func TestPollTranslatesEscapeToQuit(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	_, ok := nextEvent(t, ui).(outbound.QuitEvent)
	assert.True(t, ok, "expected QuitEvent")
}

// --- translateEvent (exercised via poll) ---

func TestPollTranslatesCtrlCToQuit(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyCtrlC, 0, tcell.ModNone)
	_, ok := nextEvent(t, ui).(outbound.QuitEvent)
	assert.True(t, ok, "expected QuitEvent")
}

func TestPollTranslatesQToQuit(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	_, ok := nextEvent(t, ui).(outbound.QuitEvent)
	assert.True(t, ok, "expected QuitEvent")
}

func TestPollTranslatesUppercaseQToQuit(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, 'Q', tcell.ModNone)
	_, ok := nextEvent(t, ui).(outbound.QuitEvent)
	assert.True(t, ok, "expected QuitEvent")
}

func TestPollTranslatesYToConfirmAccept(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, 'y', tcell.ModNone)
	ce, ok := nextEvent(t, ui).(outbound.ConfirmEvent)
	require.True(t, ok, "expected ConfirmEvent")
	assert.True(t, ce.Accept)
}

func TestPollTranslatesUppercaseYToConfirmAccept(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, 'Y', tcell.ModNone)
	ce, ok := nextEvent(t, ui).(outbound.ConfirmEvent)
	require.True(t, ok, "expected ConfirmEvent")
	assert.True(t, ce.Accept)
}

func TestPollTranslatesNToConfirmDecline(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, 'n', tcell.ModNone)
	ce, ok := nextEvent(t, ui).(outbound.ConfirmEvent)
	require.True(t, ok, "expected ConfirmEvent")
	assert.False(t, ce.Accept)
}

func TestPollTranslatesUppercaseNToConfirmDecline(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, 'N', tcell.ModNone)
	ce, ok := nextEvent(t, ui).(outbound.ConfirmEvent)
	require.True(t, ok, "expected ConfirmEvent")
	assert.False(t, ce.Accept)
}

func TestPollTranslatesPlusToSpeedFaster(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, '+', tcell.ModNone)
	se, ok := nextEvent(t, ui).(outbound.SpeedEvent)
	require.True(t, ok, "expected SpeedEvent")
	assert.True(t, se.Faster)
}

func TestPollTranslatesEqualsToSpeedFaster(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, '=', tcell.ModNone)
	se, ok := nextEvent(t, ui).(outbound.SpeedEvent)
	require.True(t, ok, "expected SpeedEvent")
	assert.True(t, se.Faster)
}

func TestPollTranslatesMinusToSpeedSlower(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, '-', tcell.ModNone)
	se, ok := nextEvent(t, ui).(outbound.SpeedEvent)
	require.True(t, ok, "expected SpeedEvent")
	assert.False(t, se.Faster)
}

func TestPollTranslatesUnderscoreToSpeedSlower(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, '_', tcell.ModNone)
	se, ok := nextEvent(t, ui).(outbound.SpeedEvent)
	require.True(t, ok, "expected SpeedEvent")
	assert.False(t, se.Faster)
}

func TestPollUnrecognizedRuneDropped(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, 'z', tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	_, ok := nextEvent(t, ui).(outbound.QuitEvent)
	assert.True(t, ok, "expected 'z' to be dropped and 'q' to translate to QuitEvent")
}

func TestPollMouseButton1EmitsClickEvent(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectMouse(5, 10, tcell.Button1, tcell.ModNone)
	ce, ok := nextEvent(t, ui).(outbound.ClickEvent)
	require.True(t, ok, "expected ClickEvent")
	assert.Equal(t, 5, ce.X)
	assert.Equal(t, 10, ce.Y)
}

// --- poll: event-type routing ---

func TestPollNonButton1DropsEvent(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectMouse(5, 10, tcell.Button2, tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	ev := nextEvent(t, ui)
	_, isClick := ev.(outbound.ClickEvent)
	assert.False(t, isClick, "Button2 should not produce a ClickEvent")
	_, ok := ev.(outbound.QuitEvent)
	require.True(t, ok, "expected QuitEvent after dropped Button2")
}

func TestPollDropsClickOnHUDRow(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectMouse(5, 0, tcell.Button1, tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	ev := nextEvent(t, ui)
	_, isClick := ev.(outbound.ClickEvent)
	assert.False(t, isClick, "a click on row 0 (the HUD row) should not produce a ClickEvent")
	_, ok := ev.(outbound.QuitEvent)
	require.True(t, ok, "expected QuitEvent after the dropped HUD-row click")
}

func TestPollDropsClickOnStatusBarRow(t *testing.T) {
	ui, screen := newUI(t)
	_, h := screen.Size()
	screen.InjectMouse(5, h-1, tcell.Button1, tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	ev := nextEvent(t, ui)
	_, isClick := ev.(outbound.ClickEvent)
	assert.False(t, isClick, "a click on the status bar row should not produce a ClickEvent")
	_, ok := ev.(outbound.QuitEvent)
	require.True(t, ok, "expected QuitEvent after the dropped status-bar-row click")
}

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

// --- drawStatusBar ---

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

func TestDrawStatusBarConfirmingMultiByteName(t *testing.T) {
	ui, screen := newUI(t)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{
			Confirming: &outbound.ConfirmViewState{PID: 42, Name: "café-server"},
		},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.Contains(t, got, "café-server", "status bar should render the multi-byte name intact, with no byte-offset gap")
	assert.Contains(t, got, "(Y)es", "status bar should still contain the confirm options after a multi-byte name")
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
	for x := range w {
		if r := cells[row*w+x].Runes; len(r) > 0 {
			sb.WriteRune(r[0])
		}
	}
	return strings.TrimRight(sb.String(), " ")
}
