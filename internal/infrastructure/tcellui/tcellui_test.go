package tcellui_test

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/core/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
)

// newUI creates an initialised UI backed by a simulation screen and registers
// Cleanup to call ui.Cleanup when the test ends.
func newUI(t *testing.T) (*tcellui.UI, tcell.SimulationScreen) {
	t.Helper()
	screen := tcell.NewSimulationScreen("")
	screen.SetSize(80, 25)
	ui := tcellui.New(screen)
	if err := ui.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ui.Cleanup)
	return ui, screen
}

// nextEvent reads one event from the UI with a timeout so tests fail fast
// instead of blocking forever if the expected event is never produced.
func nextEvent(t *testing.T, ui *tcellui.UI) event.InputEvent {
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
	ui := tcellui.New(screen)
	before := runtime.NumGoroutine()

	if err := ui.Init(); err != nil {
		t.Fatal(err)
	}

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
func TestDrawHUD_NarrowTerminalSuppressesCenter(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.New(screen)
	if err := ui.Init(); err != nil {
		t.Fatal(err)
	}
	defer ui.Cleanup()

	screen.SetSize(30, 25)
	ui.Render(game.Frame{HUD: game.HUDState{}})

	cells, w, _ := screen.GetContents()
	var row0 strings.Builder
	for x := 0; x < w; x++ {
		if r := cells[x].Runes; len(r) > 0 {
			row0.WriteRune(r[0])
		}
	}
	got := row0.String()

	if !strings.Contains(got, "FREED") {
		t.Errorf("row 0 should contain FREED: %q", got)
	}
	if !strings.Contains(got, "KILLS") {
		t.Errorf("row 0 should contain KILLS: %q", got)
	}
	if strings.Contains(got, "Highscore") {
		t.Errorf("row 0 should NOT contain Highscore on narrow terminal: %q", got)
	}
}

// TestDrawHUD_WideTerminalDrawsAllThree verifies all three HUD elements are visible on wide terminals.
func TestDrawHUD_WideTerminalDrawsAllThree(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.New(screen)
	if err := ui.Init(); err != nil {
		t.Fatal(err)
	}
	defer ui.Cleanup()

	ui.Render(game.Frame{HUD: game.HUDState{}})

	cells, w, _ := screen.GetContents()
	var row0 strings.Builder
	for x := 0; x < w; x++ {
		if r := cells[x].Runes; len(r) > 0 {
			row0.WriteRune(r[0])
		}
	}
	got := row0.String()

	for _, want := range []string{"FREED", "Highscore", "KILLS"} {
		if !strings.Contains(got, want) {
			t.Errorf("row 0 should contain %q on wide terminal: %q", want, got)
		}
	}
}

// TestRender_MultiByteLabel_ColumnLayout is a regression test for issue #13.
func TestRender_MultiByteLabel_ColumnLayout(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.New(screen)
	if err := ui.Init(); err != nil {
		t.Fatal(err)
	}
	defer ui.Cleanup()

	ui.Render(game.Frame{
		Targets: []game.TargetView{
			{X: 0, Y: 2, Label: "✦ KILLED ✦", Killing: true},
		},
	})

	cells, w, _ := screen.GetContents()
	if got := cells[2*w+1].Runes[0]; got != ' ' {
		t.Errorf("col 1 should be space (rune after ✦), got %q — byte-offset bug in render loop?", got)
	}
	if got := cells[2*w+2].Runes[0]; got != 'K' {
		t.Errorf("col 2 should be 'K', got %q — byte-offset bug in render loop?", got)
	}
}

// --- translateKey (exercised via poll) ---

func TestPoll_TranslatesEscape(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	ke, ok := nextEvent(t, ui).(event.KeyEvent)
	if !ok {
		t.Fatal("expected KeyEvent")
	}
	if ke.Key != event.KeyEscape {
		t.Errorf("expected KeyEscape, got %v", ke.Key)
	}
}

func TestPoll_TranslatesCtrlC(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyCtrlC, 0, tcell.ModNone)
	ke, ok := nextEvent(t, ui).(event.KeyEvent)
	if !ok {
		t.Fatal("expected KeyEvent")
	}
	if ke.Key != event.KeyCtrlC {
		t.Errorf("expected KeyCtrlC, got %v", ke.Key)
	}
}

func TestPoll_TranslatesCtrlZ(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyCtrlZ, 0, tcell.ModNone)
	ke, ok := nextEvent(t, ui).(event.KeyEvent)
	if !ok {
		t.Fatal("expected KeyEvent")
	}
	if ke.Key != event.KeyCtrlZ {
		t.Errorf("expected KeyCtrlZ, got %v", ke.Key)
	}
}

func TestPoll_TranslatesRune(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	ke, ok := nextEvent(t, ui).(event.KeyEvent)
	if !ok {
		t.Fatal("expected KeyEvent")
	}
	if ke.Key != event.KeyNone {
		t.Errorf("expected KeyNone for plain rune, got %v", ke.Key)
	}
	if ke.Ch != 'q' {
		t.Errorf("expected Ch='q', got %q", ke.Ch)
	}
}

// --- poll: event-type routing ---

func TestPoll_MouseButton1_EmitsClickEvent(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectMouse(5, 10, tcell.Button1, tcell.ModNone)
	ce, ok := nextEvent(t, ui).(event.ClickEvent)
	if !ok {
		t.Fatal("expected ClickEvent")
	}
	if ce.X != 5 || ce.Y != 10 {
		t.Errorf("expected ClickEvent{5,10}, got {%d,%d}", ce.X, ce.Y)
	}
}

func TestPoll_NonButton1_DropsEvent(t *testing.T) {
	ui, screen := newUI(t)
	screen.InjectMouse(5, 10, tcell.Button2, tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'z', tcell.ModNone)
	ev := nextEvent(t, ui)
	if _, isClick := ev.(event.ClickEvent); isClick {
		t.Error("Button2 should not produce a ClickEvent")
	}
	ke, ok := ev.(event.KeyEvent)
	if !ok || ke.Ch != 'z' {
		t.Errorf("expected KeyEvent{Ch:'z'} after dropped Button2, got %T %v", ev, ev)
	}
}

func TestPoll_ResizeEvent_EmitsResizeEvent(t *testing.T) {
	ui, screen := newUI(t)
	screen.PostEvent(tcell.NewEventResize(100, 40))
	if _, ok := nextEvent(t, ui).(event.ResizeEvent); !ok {
		t.Error("expected ResizeEvent from screen resize")
	}
}

// --- drawStatusBar ---

func TestDrawStatusBar_Normal(t *testing.T) {
	ui, screen := newUI(t)
	ui.Render(game.Frame{
		StatusBar: game.StatusState{Alive: 3, Speed: 2.0},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	for _, want := range []string{"Targets:", "Speed:", "Click to kill"} {
		if !strings.Contains(got, want) {
			t.Errorf("status bar should contain %q, got: %q", want, got)
		}
	}
}

func TestDrawStatusBar_Confirming(t *testing.T) {
	ui, screen := newUI(t)
	ui.Render(game.Frame{
		StatusBar: game.StatusState{
			Confirming: &game.ConfirmState{PID: 42, Name: "myapp"},
		},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	for _, want := range []string{"42", "myapp", "(Y)es", "(N)o"} {
		if !strings.Contains(got, want) {
			t.Errorf("confirm bar should contain %q, got: %q", want, got)
		}
	}
}

func TestDrawStatusBar_WithTimeLimit(t *testing.T) {
	ui, screen := newUI(t)
	ui.Render(game.Frame{
		StatusBar: game.StatusState{Alive: 1, Speed: 1.0, TimeLimit: 30, TimeLeft: 15},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	if !strings.Contains(got, "Time:") {
		t.Errorf("status bar with time limit should contain 'Time:', got: %q", got)
	}
	if !strings.Contains(got, "15s") {
		t.Errorf("status bar should show time left, got: %q", got)
	}
}

func TestDrawStatusBar_NoTimeLimit(t *testing.T) {
	ui, screen := newUI(t)
	ui.Render(game.Frame{
		StatusBar: game.StatusState{Alive: 1, Speed: 1.0, TimeLimit: 0},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	if strings.Contains(got, "Time:") {
		t.Errorf("status bar with no time limit should not contain 'Time:', got: %q", got)
	}
}
