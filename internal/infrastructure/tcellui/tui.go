// Package tcellui implements the outbound.Renderer and outbound.InputSource ports using tcell.
package tcellui

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// TUI implements both outbound.Renderer and outbound.InputSource, composing a
// renderer and a poller that share the same tcell.Screen.
type TUI struct {
	screen      tcell.Screen
	renderer    renderer
	poller      poller
	cleanupOnce sync.Once
	initialized bool
}

// Compile-time assertions that *TUI satisfies both ports, so a signature
// drift fails the build here instead of at a distant call site.
var (
	_ outbound.Renderer    = (*TUI)(nil)
	_ outbound.InputSource = (*TUI)(nil)
)

// chromeSize is the space tcellui reserves at the top and bottom of the
// display for its own fixed UI: one row for the HUD, one for the status bar.
var chromeSize = outbound.ChromeSize{Top: 1, Bottom: 1}

// NewTUI returns a tcellui.TUI wrapping the given screen.
// The caller must not call tcell.Screen.Init directly; use TUI.Init() instead.
func NewTUI(screen tcell.Screen) *TUI {
	if screen == nil {
		panic("NewTUI requires a non-nil tcell.Screen")
	}
	return &TUI{
		screen:   screen,
		renderer: newRenderer(screen),
		poller:   newPoller(screen),
	}
}

// Init initialises the screen and starts the event polling goroutine.
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

// Cleanup signals the poll goroutine to stop, then shuts down the screen.
// Safe to call more than once; only the first call has any effect. Safe to
// call even if Init was never called or failed.
func (t *TUI) Cleanup() {
	t.cleanupOnce.Do(t.cleanup)
}

// WindowSize returns the current terminal dimensions.
func (t *TUI) WindowSize() outbound.WindowSize {
	width, height := t.screen.Size()
	return outbound.WindowSize{Width: width, Height: height}
}

// ChromeSize reports that tcellui reserves one row at the top for the
// HUD and one row at the bottom for the status bar.
func (t *TUI) ChromeSize() outbound.ChromeSize {
	return chromeSize
}

// Render translates an outbound.FrameState into tcell draw calls.
func (t *TUI) Render(state outbound.FrameState) {
	t.renderer.render(state)
}

// Events returns the channel of translated game input events.
func (t *TUI) Events() <-chan outbound.InputEvent {
	return t.poller.events()
}

func (t *TUI) cleanup() {
	t.poller.stop()
	if t.initialized {
		t.screen.Fini()
	}
}

// isChromeRow reports whether row is one of the rows chrome reserves for the
// renderer's own fixed UI, given a display height rows tall.
func isChromeRow(row, height int, chrome outbound.ChromeSize) bool {
	return row < chrome.Top || row >= height-chrome.Bottom
}
