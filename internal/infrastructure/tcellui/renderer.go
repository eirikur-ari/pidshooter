package tcellui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// renderer draws one outbound.FrameViewState onto a tcell.Screen.
type renderer struct {
	screen tcell.Screen
	chrome outbound.ChromeSize
	state  outbound.FrameViewState
}

// newRenderer returns a renderer that draws state onto screen, leaving the
// rows reserved by chrome to the HUD and status bar.
func newRenderer(screen tcell.Screen, chrome outbound.ChromeSize, state outbound.FrameViewState) renderer {
	return renderer{screen: screen, chrome: chrome, state: state}
}

// render draws the frame state onto the screen.
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

// drawFrame draws the frame state within window.
func (r *renderer) drawFrame(window outbound.WindowSize) {
	for _, targetState := range r.state.Targets {
		target := target{state: targetState, window: window, chrome: r.chrome}
		target.draw(r.screen)
	}

	hud := hud{state: r.state.HUD, width: window.Width}
	hud.draw(r.screen)

	statusBar := statusBar{state: r.state.StatusBar, window: window}
	statusBar.draw(r.screen)
}

// endFrame presents the drawn frame.
func (r *renderer) endFrame() {
	r.screen.Show()
}
