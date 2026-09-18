package tcellui

import (
	"fmt"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

var statusBarStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)

// statusBar is the status bar row: the play status or, while a kill is
// pending confirmation, the confirmation prompt.
type statusBar struct {
	state  outbound.StatusState
	window outbound.WindowSize
}

// draw draws s at its window's bottom row.
func (s *statusBar) draw(screen tcell.Screen) {
	s.clear(screen)
	s.left(screen, s.text(), statusBarStyle)
}

// clear blanks s's row, ready for new text.
func (s *statusBar) clear(screen tcell.Screen) {
	row := s.window.Height - 1
	for x := range s.window.Width {
		screen.SetContent(x, row, ' ', nil, statusBarStyle)
	}
}

// left draws text left-anchored on s's row.
func (s *statusBar) left(screen tcell.Screen, text string, style tcell.Style) {
	screen.PutStrStyled(0, s.window.Height-1, text, style)
}

// text returns s's text, budgeted to fit within its window's width.
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
	if s.state.TimeLimit > 0 {
		timerStr = fmt.Sprintf(" | Time: %ds", s.state.TimeLeft)
	}
	return fmt.Sprintf(" Targets: %d | Speed: %.1fx%s | Click to kill | +/- speed | 'q' quit",
		s.state.Alive, s.state.Speed, timerStr)
}

// truncate returns text truncated to whatever room is left in s's window
// after prefix and suffix, with a trailing "…" in place of the last rune
// if text was cut.
func (s *statusBar) truncate(text, prefix, suffix string) string {
	maxRunes := s.window.Width - utf8.RuneCountInString(prefix) - utf8.RuneCountInString(suffix)
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	if maxRunes == 1 {
		return "…"
	}
	return string(runes[:maxRunes-1]) + "…"
}
