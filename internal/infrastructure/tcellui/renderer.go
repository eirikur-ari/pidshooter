package tcellui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// renderer draws one outbound.FrameState onto a tcell.Screen.
type renderer struct {
	screen tcell.Screen
	chrome outbound.ChromeSize
	state  outbound.FrameState
}

// newRenderer returns a renderer that will draw state onto screen,
// treating chrome's rows as reserved for the caller's own fixed UI.
func newRenderer(screen tcell.Screen, chrome outbound.ChromeSize, state outbound.FrameState) renderer {
	return renderer{screen: screen, chrome: chrome, state: state}
}

// render draws r's state: clears the screen, draws targets, HUD, and
// status bar, then presents the result.
func (r *renderer) render() {
	window := r.beginFrame()
	r.drawFrame(window)
	r.endFrame()
}

// beginFrame clears the screen and returns the current window size, ready
// for a new frame's draw calls.
func (r *renderer) beginFrame() outbound.WindowSize {
	r.screen.Clear()
	width, height := r.screen.Size()
	return outbound.WindowSize{Width: width, Height: height}
}

// drawFrame draws r's state within window.
func (r *renderer) drawFrame(window outbound.WindowSize) {
	for _, ts := range r.state.Targets {
		tg := target{state: ts, window: window, chrome: r.chrome}
		tg.draw(r.screen)
	}

	h := hud{state: r.state.HUD, width: window.Width}
	h.draw(r.screen)

	sb := statusBar{state: r.state.StatusBar, window: window}
	sb.draw(r.screen)
}

// endFrame presents the frame's draw calls to the terminal.
func (r *renderer) endFrame() {
	r.screen.Show()
}
