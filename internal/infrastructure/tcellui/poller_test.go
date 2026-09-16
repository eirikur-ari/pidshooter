package tcellui_test

import (
	"runtime"
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
	ui := tcellui.NewTUI(screen)
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

func TestEventsChannelClosesAfterCleanup(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewTUI(screen)
	require.NoError(t, ui.Init())

	ui.Cleanup()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		select {
		case _, ok := <-ui.Events():
			if !ok {
				return
			}
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
	t.Fatal("Events channel was not closed after Cleanup")
}

func TestPollTranslatesEscapeToQuit(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	_, ok := nextEvent(t, ui).(outbound.QuitEvent)
	assert.True(t, ok, "expected QuitEvent")
}

// --- translateKeyEvent (exercised via poll) ---

func TestPollTranslatesCtrlCToQuit(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyCtrlC, 0, tcell.ModNone)
	_, ok := nextEvent(t, ui).(outbound.QuitEvent)
	assert.True(t, ok, "expected QuitEvent")
}

func TestPollTranslatesCtrlZToQuit(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyCtrlZ, 0, tcell.ModNone)
	_, ok := nextEvent(t, ui).(outbound.QuitEvent)
	assert.True(t, ok, "expected QuitEvent")
}

func TestPollTranslatesQToQuit(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	_, ok := nextEvent(t, ui).(outbound.QuitEvent)
	assert.True(t, ok, "expected QuitEvent")
}

func TestPollTranslatesUppercaseQToQuit(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, 'Q', tcell.ModNone)
	_, ok := nextEvent(t, ui).(outbound.QuitEvent)
	assert.True(t, ok, "expected QuitEvent")
}

func TestPollTranslatesYToConfirmAccept(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, 'y', tcell.ModNone)
	ce, ok := nextEvent(t, ui).(outbound.ConfirmEvent)
	require.True(t, ok, "expected ConfirmEvent")
	assert.True(t, ce.Accept)
}

func TestPollTranslatesUppercaseYToConfirmAccept(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, 'Y', tcell.ModNone)
	ce, ok := nextEvent(t, ui).(outbound.ConfirmEvent)
	require.True(t, ok, "expected ConfirmEvent")
	assert.True(t, ce.Accept)
}

func TestPollTranslatesNToConfirmDecline(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, 'n', tcell.ModNone)
	ce, ok := nextEvent(t, ui).(outbound.ConfirmEvent)
	require.True(t, ok, "expected ConfirmEvent")
	assert.False(t, ce.Accept)
}

func TestPollTranslatesUppercaseNToConfirmDecline(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, 'N', tcell.ModNone)
	ce, ok := nextEvent(t, ui).(outbound.ConfirmEvent)
	require.True(t, ok, "expected ConfirmEvent")
	assert.False(t, ce.Accept)
}

func TestPollTranslatesPlusToSpeedFaster(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, '+', tcell.ModNone)
	se, ok := nextEvent(t, ui).(outbound.SpeedEvent)
	require.True(t, ok, "expected SpeedEvent")
	assert.True(t, se.Faster)
}

func TestPollTranslatesEqualsToSpeedFaster(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, '=', tcell.ModNone)
	se, ok := nextEvent(t, ui).(outbound.SpeedEvent)
	require.True(t, ok, "expected SpeedEvent")
	assert.True(t, se.Faster)
}

func TestPollTranslatesMinusToSpeedSlower(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, '-', tcell.ModNone)
	se, ok := nextEvent(t, ui).(outbound.SpeedEvent)
	require.True(t, ok, "expected SpeedEvent")
	assert.False(t, se.Faster)
}

func TestPollTranslatesUnderscoreToSpeedSlower(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, '_', tcell.ModNone)
	se, ok := nextEvent(t, ui).(outbound.SpeedEvent)
	require.True(t, ok, "expected SpeedEvent")
	assert.False(t, se.Faster)
}

func TestPollUnrecognizedRuneDropped(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, 'z', tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	_, ok := nextEvent(t, ui).(outbound.QuitEvent)
	assert.True(t, ok, "expected 'z' to be dropped and 'q' to translate to QuitEvent")
}

func TestPollMouseButton1EmitsClickEvent(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectMouse(5, 10, tcell.Button1, tcell.ModNone)
	ce, ok := nextEvent(t, ui).(outbound.ClickEvent)
	require.True(t, ok, "expected ClickEvent")
	assert.Equal(t, 5, ce.X)
	assert.Equal(t, 10, ce.Y)
}

// --- poll: event-type routing ---

func TestPollNonButton1DropsEvent(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectMouse(5, 10, tcell.Button2, tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	ev := nextEvent(t, ui)
	_, isClick := ev.(outbound.ClickEvent)
	assert.False(t, isClick, "Button2 should not produce a ClickEvent")
	_, ok := ev.(outbound.QuitEvent)
	require.True(t, ok, "expected QuitEvent after dropped Button2")
}

func TestPollDropsClickOnHUDRow(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectMouse(5, 0, tcell.Button1, tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	ev := nextEvent(t, ui)
	_, isClick := ev.(outbound.ClickEvent)
	assert.False(t, isClick, "a click on row 0 (the HUD row) should not produce a ClickEvent")
	_, ok := ev.(outbound.QuitEvent)
	require.True(t, ok, "expected QuitEvent after the dropped HUD-row click")
}

func TestPollDropsClickOnStatusBarRow(t *testing.T) {
	ui, screen := newTUI(t)
	_, h := screen.Size()
	screen.InjectMouse(5, h-1, tcell.Button1, tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	ev := nextEvent(t, ui)
	_, isClick := ev.(outbound.ClickEvent)
	assert.False(t, isClick, "a click on the status bar row should not produce a ClickEvent")
	_, ok := ev.(outbound.QuitEvent)
	require.True(t, ok, "expected QuitEvent after the dropped status-bar-row click")
}

// nextEvent reads one event from the UI with a timeout so tests fail fast
// instead of blocking forever if the expected event is never produced.
func nextEvent(t *testing.T, ui *tcellui.TUI) outbound.InputEvent {
	t.Helper()
	select {
	case ev := <-ui.Events():
		return ev
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for event from poll goroutine")
		return nil
	}
}
