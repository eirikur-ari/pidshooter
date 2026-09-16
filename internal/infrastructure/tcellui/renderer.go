package tcellui

import (
	"fmt"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// renderer draws game state to a tcell.Screen.
type renderer struct {
	screen tcell.Screen
}

var (
	aliveStyle     = tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	killStyle      = tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	fleeStyle      = tcell.StyleDefault.Foreground(tcell.ColorOrange).Bold(true)
	freedStyle     = tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true)
	highScoreStyle = tcell.StyleDefault.Foreground(tcell.ColorPurple).Bold(true)
	killsStyle     = tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	statusBarStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
)

// newRenderer returns a renderer drawing to screen.
func newRenderer(screen tcell.Screen) renderer {
	return renderer{screen: screen}
}

// render translates an outbound.FrameState into tcell draw calls.
func (r *renderer) render(state outbound.FrameState) {
	r.screen.Clear()
	width, height := r.screen.Size()
	window := outbound.WindowSize{Width: width, Height: height}

	for _, target := range state.Targets {
		r.drawTarget(target, window, chromeSize)
	}

	r.drawHUD(window.Width, state.HUD)
	r.drawStatusBar(window, state.StatusBar)
	r.screen.Show()
}

// drawTarget draws target's tag or animation frame, unless its position
// falls outside window or on one of chrome's reserved rows.
func (r *renderer) drawTarget(target outbound.TargetViewState, window outbound.WindowSize, chrome outbound.ChromeSize) {
	if target.X < 0 || target.X >= window.Width || isChromeRow(target.Y, window.Height, chrome) {
		return
	}
	style, tag := targetStyleAndTag(target)
	r.screen.PutStrStyled(target.X, target.Y, tag, style)
}

func (r *renderer) drawHUD(width int, hud outbound.HUDState) {
	freedStr := fmt.Sprintf(" FREED: %s ", util.FormatBytes(hud.FreedMem))
	freedWidth := utf8.RuneCountInString(freedStr)

	highScoreStr := fmt.Sprintf(" Highscore: %d ", hud.HighScore)
	highScoreWidth := utf8.RuneCountInString(highScoreStr)
	highScoreX := (width - highScoreWidth) / 2

	killsStr := fmt.Sprintf(" KILLS: %d ", hud.Kills)
	killsWidth := utf8.RuneCountInString(killsStr)
	killsX := max(width-killsWidth, 0)

	r.screen.PutStrStyled(0, 0, truncateRunes(freedStr, killsX), freedStyle)

	if highScoreX >= freedWidth && highScoreX+highScoreWidth <= killsX {
		r.screen.PutStrStyled(highScoreX, 0, highScoreStr, highScoreStyle)
	}

	r.screen.PutStrStyled(killsX, 0, killsStr, killsStyle)
}

func (r *renderer) drawStatusBar(window outbound.WindowSize, status outbound.StatusState) {
	row := window.Height - 1
	for x := range window.Width {
		r.screen.SetContent(x, row, ' ', nil, statusBarStyle)
	}
	r.screen.PutStrStyled(0, row, statusBarText(status, window.Width), statusBarStyle)
}

// targetStyleAndTag returns the style and label to draw for target,
// substituting the current animation frame while Killing or Fleeing.
func targetStyleAndTag(target outbound.TargetViewState) (tcell.Style, string) {
	switch {
	case target.Killing:
		return killStyle, killAnimation.frame(target.AnimationProgress)
	case target.Fleeing:
		return fleeStyle, fleeAnimation.frame(target.AnimationProgress)
	default:
		return aliveStyle, target.Tag
	}
}

// statusBarText returns the status bar's text for status, budgeted to fit
// within width columns.
func statusBarText(status outbound.StatusState, width int) string {
	if status.Confirming != nil {
		return confirmPromptText(status.Confirming, width)
	}
	return playStatusText(status)
}

// confirmPromptText returns the kill-confirmation prompt, truncating the
// target's name with an ellipsis if it would overflow width.
func confirmPromptText(confirming *outbound.ConfirmViewState, width int) string {
	prefix := fmt.Sprintf(" Kill [%d ", confirming.PID)
	const suffix = "]? (Y)es / (N)o / (Q)uit"
	nameBudget := width - utf8.RuneCountInString(prefix) - utf8.RuneCountInString(suffix)
	return prefix + truncateWithEllipsis(confirming.Name, nameBudget) + suffix
}

// playStatusText returns the normal (non-confirming) status line.
func playStatusText(status outbound.StatusState) string {
	timerStr := ""
	if status.TimeLimit > 0 {
		timerStr = fmt.Sprintf(" | Time: %ds", status.TimeLeft)
	}
	return fmt.Sprintf(" Targets: %d | Speed: %.1fx%s | Click to kill | +/- speed | 'q' quit",
		status.Alive, status.Speed, timerStr)
}

// truncateRunes returns s truncated to at most n runes, so a caller can give
// PutStrStyled an explicit right-edge budget instead of relying on a
// later draw call to overwrite whatever runs past it.
func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

// truncateWithEllipsis is truncateRunes for text a player reads directly,
// signaling with a trailing "…" that something was cut rather than
// silently dropping it.
func truncateWithEllipsis(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n == 1 {
		return "…"
	}
	return string(runes[:n-1]) + "…"
}
