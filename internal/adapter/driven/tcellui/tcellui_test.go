package tcellui_test

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/adapter/driven/tcellui"
	driven "github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"
)

// TestPollGoroutineExitsAfterCleanup is a regression test for issue #6.
// It verifies that the poll goroutine terminates after Cleanup() even when it
// is blocked on a channel send (not on PollEvent). Before the fix, poll had no
// way to unblock from the send once the game loop stopped consuming events.
func TestPollGoroutineExitsAfterCleanup(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.New(screen)
	before := runtime.NumGoroutine()

	if err := ui.Init(); err != nil {
		t.Fatal(err)
	}

	// Inject more events than the channel buffer capacity (10) without consuming
	// any of them. poll will fill ch then block on the 11th channel send,
	// reproducing the scenario where the game loop has stopped draining events.
	for i := 0; i < 15; i++ {
		screen.InjectKey(tcell.KeyRune, 'a', tcell.ModNone)
	}
	time.Sleep(50 * time.Millisecond)

	ui.Cleanup()

	// Poll until the goroutine count returns to baseline or the deadline expires.
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
// It verifies that on a narrow terminal where the Highscore label would overlap
// the FREED or KILLS labels, the center element is suppressed rather than drawn
// on top of the outer elements.
func TestDrawHUD_NarrowTerminalSuppressesCenter(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.New(screen)
	if err := ui.Init(); err != nil {
		t.Fatal(err)
	}
	defer ui.Cleanup()

	// At width=30 with zero values:
	//   FREED  label is 13 chars (cols 0–12)
	//   hiX = (30–14)/2 = 8, which is inside FREED → guard must suppress Highscore
	//   KILLS  label is 11 chars, starts at col 19
	screen.SetSize(30, 25)
	ui.Render(driven.Frame{HUD: driven.HUDState{}})

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

// TestDrawHUD_WideTerminalDrawsAllThree verifies that on a wide terminal all three
// HUD elements are visible simultaneously.
func TestDrawHUD_WideTerminalDrawsAllThree(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.New(screen)
	if err := ui.Init(); err != nil {
		t.Fatal(err)
	}
	defer ui.Cleanup()

	// Default 80×25 simulation screen — all three elements fit without overlap.
	ui.Render(driven.Frame{HUD: driven.HUDState{}})

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
// It verifies that multi-byte characters in kill-animation labels render at the
// correct screen columns. Before the fix, the rendering loop used the byte offset
// from range as the column index, so multi-byte runes shifted all subsequent
// characters right by (byteLen - 1) extra columns.
func TestRender_MultiByteLabel_ColumnLayout(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.New(screen)
	if err := ui.Init(); err != nil {
		t.Fatal(err)
	}
	defer ui.Cleanup()

	// "✦ KILLED ✦": ✦ is 3 UTF-8 bytes. With the byte-offset bug, the space after ✦
	// lands at column 3 instead of column 1, and 'K' lands at column 4 instead of 2.
	ui.Render(driven.Frame{
		Targets: []driven.TargetView{
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
