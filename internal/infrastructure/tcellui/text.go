package tcellui

import "github.com/rivo/uniseg"

// truncateToWidth returns the longest prefix of text that occupies at most
// maxWidth terminal columns, never splitting a character.
func truncateToWidth(text string, maxWidth int) string {
	width := 0
	end := 0
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		width += graphemes.Width()
		if width > maxWidth {
			break
		}
		end += len(graphemes.Str())
	}
	return text[:end]
}
