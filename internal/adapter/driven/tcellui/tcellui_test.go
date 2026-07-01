package tcellui_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/adapter/driven/tcellui"
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
