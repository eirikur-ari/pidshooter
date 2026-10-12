package tcellui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

func screenRow(screen tcell.SimulationScreen, row int) string {
	screen.Show()
	cells, width, _ := screen.GetContents()
	var content strings.Builder
	nextColumn := 0
	for column := range width {
		// A wide character covers the cells after its own; they hold no runes.
		if column < nextColumn {
			continue
		}
		runes := cells[row*width+column].Runes
		if len(runes) == 0 {
			content.WriteRune(' ')
			continue
		}
		content.WriteRune(runes[0])
		nextColumn = column + uniseg.StringWidth(string(runes[0]))
	}
	return strings.TrimRight(content.String(), " ")
}
