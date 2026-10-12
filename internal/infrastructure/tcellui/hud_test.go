package tcellui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestHUD_draw_ArrangesLabelsByWindowWidth(t *testing.T) {
	tests := newHUDLayoutTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			screen := newScreenFixture(test.width, 25)
			hud := hud{state: outbound.HUDViewState{Kills: 7}, width: test.width}

			// When
			hud.draw(screen)

			// Then
			assert.Equal(t, test.expected, screenRow(screen, 0))
		})
	}
}

func TestHUD_draw_ShowsEveryValue(t *testing.T) {
	// Given
	screen := newScreenFixture(80, 25)
	hud := hud{
		state: outbound.HUDViewState{FreedMem: 2048, Kills: 7, HighScore: 12},
		width: 80,
	}

	// When
	hud.draw(screen)

	// Then
	row := screenRow(screen, 0)
	assert.Contains(t, row, " FREED: 2.0 KB ")
	assert.Contains(t, row, " HighScore: 12 ")
	assert.Contains(t, row, " KILLS: 7")
}

func TestHUD_left_DrawsLongestPrefixWithinRightBound(t *testing.T) {
	tests := newHUDLeftTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			screen := newScreenFixture(20, 25)
			hud := hud{bounds: bounds{right: test.right}}

			// When
			hud.left(screen, test.text, tcell.StyleDefault)

			// Then
			assert.Equal(t, test.expected, screenRow(screen, 0))
		})
	}
}

func newHUDLayoutTestCase() []struct {
	name     string
	width    int
	expected string
} {
	return []struct {
		name     string
		width    int
		expected string
	}{
		{"freed label is cut to the kills label's budget", 14, " FRE KILLS: 7"},
		{"freed label is cut to nothing when the kills label fills the row", 10, " KILLS: 7"},
		{"freed label fits whole up to the kills label's budget", 18, " FREED:  KILLS: 7"},
		{"high score is dropped when one column short of clearing the freed label", 36, " FREED: 0 B                KILLS: 7"},
		{"high score is drawn when it exactly touches the freed label", 38, " FREED: 0 B  HighScore: 0    KILLS: 7"},
	}
}

func newHUDLeftTestCase() []struct {
	name     string
	text     string
	right    int
	expected string
} {
	return []struct {
		name     string
		text     string
		right    int
		expected string
	}{
		{"text narrower than the bound is unchanged", "abc", 5, "abc"},
		{"text exactly the bound is unchanged", "abc", 3, "abc"},
		{"text wider than the bound is cut", "abcdef", 4, "abcd"},
		{"empty text stays empty", "", 4, ""},
		{"zero bound yields nothing", "abc", 0, ""},
		{"negative bound yields nothing", "abc", -2, ""},
		{"wide characters count two columns each", "日本語", 4, "日本"},
		{"a wide character that does not fit is dropped whole", "日本語", 3, "日"},
		{"a wide first character wider than the bound yields nothing", "日本語", 1, ""},
		{"a combining mark stays with its base character", "éx", 1, "é"},
	}
}
