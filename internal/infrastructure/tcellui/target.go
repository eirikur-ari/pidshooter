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

// target is a single target to draw, together with the window and chrome
// it is drawn within.
type target struct {
	state  outbound.TargetViewState
	window outbound.WindowSize
	chrome outbound.ChromeSize
}

// draw draws the target's tag or animation frame, unless the target lies
// outside the window or on a row reserved by chrome.
func (t *target) draw(screen tcell.Screen) {
	if t.state.X < 0 || t.state.X >= t.window.Width || t.isChromeRow() {
		return
	}
	style, tag := t.styleAndTag()
	screen.PutStrStyled(t.state.X, t.state.Y, tag, style)
}

// isChromeRow reports whether the target's row is reserved by chrome.
func (t *target) isChromeRow() bool {
	return t.state.Y < t.chrome.Top || t.state.Y >= t.window.Height-t.chrome.Bottom
}

// styleAndTag returns the style and label to draw for the target,
// substituting the current animation frame while Killing or Fleeing.
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
