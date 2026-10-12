package tcellui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateToWidth_SelectsLongestPrefixWithinWidth(t *testing.T) {
	tests := newTruncateToWidthTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			truncated := truncateToWidth(test.text, test.maxWidth)

			// Then
			assert.Equal(t, test.expected, truncated)
		})
	}
}

func newTruncateToWidthTestCase() []struct {
	name     string
	text     string
	maxWidth int
	expected string
} {
	return []struct {
		name     string
		text     string
		maxWidth int
		expected string
	}{
		{"text narrower than the width is unchanged", "abc", 5, "abc"},
		{"text exactly the width is unchanged", "abc", 3, "abc"},
		{"text wider than the width is cut", "abcdef", 4, "abcd"},
		{"empty text stays empty", "", 4, ""},
		{"zero width yields nothing", "abc", 0, ""},
		{"negative width yields nothing", "abc", -2, ""},
		{"wide characters count two columns each", "日本語", 4, "日本"},
		{"a wide character that does not fit is dropped whole", "日本語", 3, "日"},
		{"a wide first character wider than the width yields nothing", "日本語", 1, ""},
		{"a combining mark stays with its base character", "éx", 1, "é"},
	}
}
