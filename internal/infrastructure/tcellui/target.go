package tcellui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

var (
	aliveStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	killStyle  = tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	fleeStyle  = tcell.StyleDefault.Foreground(tcell.ColorOrange).Bold(true)
)

// target is a single target: its view state, and the window and chrome
// context needed to decide whether and how to draw it.
type target struct {
	state  outbound.TargetViewState
	window outbound.WindowSize
	chrome outbound.ChromeSize
}

// draw draws t's tag or animation frame, unless its position falls outside
// its window or on one of chrome's reserved rows.
func (t *target) draw(screen tcell.Screen) {
	if t.state.X < 0 || t.state.X >= t.window.Width || t.isChromeRow() {
		return
	}
	style, tag := t.styleAndTag()
	screen.PutStrStyled(t.state.X, t.state.Y, tag, style)
}

// isChromeRow reports whether t's row is one of the rows chrome reserves
// for the caller's own fixed UI.
func (t *target) isChromeRow() bool {
	return t.state.Y < t.chrome.Top || t.state.Y >= t.window.Height-t.chrome.Bottom
}

// styleAndTag returns the style and label to draw for t, substituting the
// current animation frame while Killing or Fleeing.
func (t *target) styleAndTag() (tcell.Style, string) {
	switch {
	case t.state.Killing:
		return killStyle, killAnimation.frame(t.state.AnimationProgress)
	case t.state.Fleeing:
		return fleeStyle, fleeAnimation.frame(t.state.AnimationProgress)
	default:
		return aliveStyle, t.state.Tag
	}
}
