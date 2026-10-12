package tcellui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

var statusBarStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)

// statusBar is the status bar row: the play status or, while a kill is
// pending confirmation, the confirmation prompt.
type statusBar struct {
	state  outbound.StatusViewState
	window outbound.WindowSize
}

// draw draws the status bar on the window's bottom row.
func (s *statusBar) draw(screen tcell.Screen) {
	s.clear(screen)
	s.left(screen, s.text(), statusBarStyle)
}

// clear blanks the status bar's row, ready for new text.
func (s *statusBar) clear(screen tcell.Screen) {
	row := s.window.Height - 1
	for x := range s.window.Width {
		screen.SetContent(x, row, ' ', nil, statusBarStyle)
	}
}

// left draws text left-anchored on the status bar's row.
func (s *statusBar) left(screen tcell.Screen, text string, style tcell.Style) {
	screen.PutStrStyled(0, s.window.Height-1, text, style)
}

// text returns the confirmation prompt while a kill is pending
// confirmation, and the play status otherwise.
func (s *statusBar) text() string {
	if s.state.Confirming != nil {
		return s.confirmPromptText()
	}
	return s.playStatusText()
}

// confirmPromptText returns the kill-confirmation prompt, truncating the
// target's name with an ellipsis if it would overflow the window's width.
func (s *statusBar) confirmPromptText() string {
	confirming := s.state.Confirming
	prefix := fmt.Sprintf(" Kill [%d ", confirming.PID)
	const suffix = "]? (Y)es / (N)o / (Q)uit"
	return prefix + s.truncate(confirming.Name, prefix, suffix) + suffix
}

// playStatusText returns the normal (non-confirming) status line.
func (s *statusBar) playStatusText() string {
	timerStr := ""
	if s.state.TimeLeft != nil {
		timerStr = fmt.Sprintf(" | Time: %ds", *s.state.TimeLeft)
	}
	return fmt.Sprintf(" Targets: %d | Speed: %.1fx%s | Click to kill | +/- speed | 'q' quit",
		s.state.Alive, s.state.Speed, timerStr)
}

// truncate returns text truncated to whatever room is left in the window
// after prefix and suffix, with a trailing "…" in place of the last column
// if text was cut.
func (s *statusBar) truncate(text, prefix, suffix string) string {
	maxWidth := s.window.Width - uniseg.StringWidth(prefix) - uniseg.StringWidth(suffix)
	if maxWidth <= 0 {
		return ""
	}
	if uniseg.StringWidth(text) <= maxWidth {
		return text
	}

	// Keep the longest prefix that leaves one column for the ellipsis,
	// never splitting a character.
	width := 0
	end := 0
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		width += graphemes.Width()
		if width > maxWidth-1 {
			break
		}
		end += len(graphemes.Str())
	}
	return text[:end] + "…"
}
