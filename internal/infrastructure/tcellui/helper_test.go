package tcellui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
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
