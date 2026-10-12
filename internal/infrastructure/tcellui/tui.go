package tcellui

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// Compile-time assertion that *TUI satisfies outbound.Renderer.
var _ outbound.Renderer = (*TUI)(nil)

// TUI renders to a tcell.Screen.
type TUI struct {
	screen      tcell.Screen
	chrome      outbound.ChromeSize
	poller      poller
	cleanupOnce sync.Once
	initialized bool
}

// NewTUI returns a TUI that draws on screen, reserving one row at the top
// for the HUD and one at the bottom for the status bar. It panics if screen
// is nil. Initialize the screen through TUI.Init, never directly.
func NewTUI(screen tcell.Screen) *TUI {
	if screen == nil {
		panic("NewTUI requires a non-nil tcell.Screen")
	}
	return &TUI{
		screen: screen,
		chrome: outbound.ChromeSize{Top: 1, Bottom: 1},
		poller: newPoller(screen),
	}
}

// Init initializes the screen and starts delivering input events. It returns
// an error if the TUI is already initialized or the screen fails to
// initialize.
func (t *TUI) Init() error {
	if t.initialized {
		return errors.New("already initialized")
	}
	if err := t.screen.Init(); err != nil {
		return fmt.Errorf("failed to initialize screen: %w", err)
	}
	t.screen.EnableMouse(tcell.MouseButtonEvents)
	t.screen.SetStyle(tcell.StyleDefault)

	t.initialized = true

	// Start the poller goroutine to read input events.
	go t.poller.poll()
	return nil
}

// Cleanup stops delivering input events and shuts down the screen. Safe to
// call more than once; only the first call has any effect. Safe to call
// even if Init was never called or failed.
func (t *TUI) Cleanup() {
	t.cleanupOnce.Do(t.cleanup)
}

// WindowSize returns the current terminal dimensions.
func (t *TUI) WindowSize() outbound.WindowSize {
	width, height := t.screen.Size()
	return outbound.WindowSize{Width: width, Height: height}
}

// ChromeSize returns the rows reserved for the HUD at the top and the
// status bar at the bottom.
func (t *TUI) ChromeSize() outbound.ChromeSize {
	return t.chrome
}

// Render draws state onto the screen.
func (t *TUI) Render(state outbound.FrameViewState) {
	r := newRenderer(t.screen, t.chrome, state)
	r.render()
}

// InputEvents returns the provider of the user's input events. Its channel
// is closed once Cleanup is called.
func (t *TUI) InputEvents() outbound.InputEventProvider {
	return &t.poller
}

// cleanup stops the poller and finalizes the screen if Init succeeded.
func (t *TUI) cleanup() {
	t.poller.stop()
	if t.initialized {
		t.screen.Fini()
	}
}
