// Package tcellui implements the outbound.Renderer and outbound.InputEventProvider ports using tcell.
package tcellui

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
)

// Compile-time assertions that *TUI satisfies both ports, so a signature
// drift fails the build here instead of at a distant call site.
var (
	_ outbound.Renderer           = (*TUI)(nil)
	_ outbound.InputEventProvider = (*TUI)(nil)
)

// TUI implements both outbound.Renderer and outbound.InputEventProvider, composing a
// poller that shares the same tcell.Screen.
type TUI struct {
	screen      tcell.Screen
	chrome      outbound.ChromeSize
	poller      poller
	cleanupOnce sync.Once
	initialized bool
}

// NewTUI returns a tcellui.TUI wrapping the given screen, reserving one row
// at the top and one row at the bottom for its own fixed UI: the HUD and
// the status bar.
// The caller must not call tcell.Screen.Init directly; use TUI.Init() instead.
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
	return t.chrome
}

// Render translates an outbound.FrameState into tcell draw calls.
func (t *TUI) Render(state outbound.FrameState) {
	r := newRenderer(t.screen, t.chrome, state)
	r.render()
}

// Events returns the channel of translated input events.
func (t *TUI) Events() <-chan input.EventDispatcher {
	return t.poller.events()
}

func (t *TUI) cleanup() {
	t.poller.stop()
	if t.initialized {
		t.screen.Fini()
	}
}
