package tcellui

import (
	"fmt"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

var (
	freedStyle     = tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true)
	highScoreStyle = tcell.StyleDefault.Foreground(tcell.ColorPurple).Bold(true)
	killsStyle     = tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
)

// bounds is the leftmost and rightmost column two neighboring HUD labels
// must not cross.
type bounds struct {
	left  int
	right int
}

// hud is the HUD row: freed memory, high score, and kill count.
type hud struct {
	state outbound.HUDState
	width int

	bounds bounds
}

// draw draws h at row 0: the kill count right-anchored, the freed-memory
// total left-anchored, and the high score centered.
func (h *hud) draw(screen tcell.Screen) {
	h.right(screen, h.killsText(), killsStyle)
	h.left(screen, h.freedText(), freedStyle)
	h.center(screen, h.highScoreText(), highScoreStyle)
}

// killsText returns the kill count's label.
func (h *hud) killsText() string {
	return fmt.Sprintf(" KILLS: %d ", h.state.Kills)
}

// freedText returns the freed-memory total's label.
func (h *hud) freedText() string {
	return fmt.Sprintf(" FREED: %s ", util.FormatBytes(h.state.FreedMem))
}

// highScoreText returns the high score's label.
func (h *hud) highScoreText() string {
	return fmt.Sprintf(" HighScore: %d ", h.state.HighScore)
}

// right draws text right-anchored, and records the column it starts at as
// bounds.right.
func (h *hud) right(screen tcell.Screen, text string, style tcell.Style) {
	width := utf8.RuneCountInString(text)
	h.bounds.right = max(h.width-width, 0)
	screen.PutStrStyled(h.bounds.right, 0, text, style)
}

// left draws text left-anchored, truncated to bounds.right, and records
// its untruncated width as bounds.left.
func (h *hud) left(screen tcell.Screen, text string, style tcell.Style) {
	h.bounds.left = utf8.RuneCountInString(text)
	screen.PutStrStyled(0, 0, h.truncate(text), style)
}

// center draws text centered, unless it would collide with bounds.left or
// bounds.right.
func (h *hud) center(screen tcell.Screen, text string, style tcell.Style) {
	width := utf8.RuneCountInString(text)
	x := (h.width - width) / 2
	if x >= h.bounds.left && x+width <= h.bounds.right {
		screen.PutStrStyled(x, 0, text, style)
	}
}

// truncate returns text truncated to at most maxRunes runes.
func (h *hud) truncate(text string) string {
	maxRunes := h.bounds.right
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes])
}
