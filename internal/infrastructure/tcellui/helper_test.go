package tcellui

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/input"
)

func screenRow(screen tcell.SimulationScreen, row int) string {
	screen.Show()
	cells, width, _ := screen.GetContents()
	var content strings.Builder
	for x := range width {
		runes := cells[row*width+x].Runes
		if len(runes) == 0 {
			content.WriteRune(' ')
			continue
		}
		content.WriteRune(runes[0])
	}
	return strings.TrimRight(content.String(), " ")
}

// isClosed drains events without blocking and reports whether it found the
// channel closed.
func isClosed(events <-chan input.Event) bool {
	for {
		select {
		case _, open := <-events:
			if !open {
				return true
			}
		default:
			return false
		}
	}
}
