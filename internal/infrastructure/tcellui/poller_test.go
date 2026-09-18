package tcellui_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/input"
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

func TestPollForwardsTranslatedEventToChannel(t *testing.T) {
	ui, screen := newTUI(t)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	assert.Equal(t, input.QuitEvent{}, nextEvent(t, ui))
}

// nextEvent reads one event from the UI with a timeout so tests fail fast
// instead of blocking forever if the expected event is never produced.
func nextEvent(t *testing.T, ui *tcellui.TUI) input.Event {
	t.Helper()
	select {
	case ev := <-ui.Events():
		return ev
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for event from poll goroutine")
		return nil
	}
}
