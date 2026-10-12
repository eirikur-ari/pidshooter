package tcellui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestTarget_draw_DrawsWithinWindowAndOutsideChrome(t *testing.T) {
	tests := newTargetDrawTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			screen := newScreenFixture(80, 25)
			target := target{
				state:  test.state,
				window: outbound.WindowSize{Width: 80, Height: 25},
				chrome: outbound.ChromeSize{Top: 1, Bottom: 1},
			}

			// When
			target.draw(screen)

			// Then
			assert.Equal(t, test.expected, screenRow(screen, test.state.Y))
		})
	}
}

func newTargetDrawTestCase() []struct {
	name     string
	state    outbound.TargetViewState
	expected string
} {
	return []struct {
		name     string
		state    outbound.TargetViewState
		expected string
	}{
		{"alive target shows its tag", outbound.TargetViewState{X: 0, Y: 5, Tag: "[1234 victim]"}, "[1234 victim]"},
		{"tag starts at the target's column", outbound.TargetViewState{X: 3, Y: 5, Tag: "[1234 victim]"}, "   [1234 victim]"},
		{"tag with multi-byte characters keeps its column layout", outbound.TargetViewState{X: 0, Y: 5, Tag: "✦ KILLED ✦"}, "✦ KILLED ✦"},
		{"tag with wide characters is drawn whole", outbound.TargetViewState{X: 0, Y: 5, Tag: "[9 日本語]"}, "[9 日本語]"},
		{"killing target shows the first kill frame instead of its tag", outbound.TargetViewState{X: 0, Y: 5, Tag: "[1234 victim]", Killing: true, AnimationProgress: 0}, "💥"},
		{"killing target shows a later kill frame as progress advances", outbound.TargetViewState{X: 0, Y: 5, Killing: true, AnimationProgress: 0.3}, "✦ KILLED ✦"},
		{"fleeing target shows the first flee frame instead of its tag", outbound.TargetViewState{X: 0, Y: 5, Tag: "[1234 victim]", Fleeing: true, AnimationProgress: 0}, "🏃💨"},
		{"fleeing target shows a later flee frame as progress advances", outbound.TargetViewState{X: 0, Y: 5, Fleeing: true, AnimationProgress: 0.3}, "↝ RAN AWAY ↝"},
		{"killing takes precedence over fleeing", outbound.TargetViewState{X: 0, Y: 5, Killing: true, Fleeing: true, AnimationProgress: 0}, "💥"},
		{"first row below the top chrome is drawn", outbound.TargetViewState{X: 0, Y: 1, Tag: "T"}, "T"},
		{"last row above the bottom chrome is drawn", outbound.TargetViewState{X: 0, Y: 23, Tag: "T"}, "T"},
		{"top chrome row is left alone", outbound.TargetViewState{X: 0, Y: 0, Tag: "T"}, ""},
		{"bottom chrome row is left alone", outbound.TargetViewState{X: 0, Y: 24, Tag: "T"}, ""},
		{"last column is drawn", outbound.TargetViewState{X: 79, Y: 5, Tag: "T"}, strings.Repeat(" ", 79) + "T"},
		{"column left of the window is not drawn", outbound.TargetViewState{X: -1, Y: 5, Tag: "T"}, ""},
		{"column right of the window is not drawn", outbound.TargetViewState{X: 80, Y: 5, Tag: "T"}, ""},
		{"tag running past the right edge is clipped", outbound.TargetViewState{X: 78, Y: 5, Tag: "[abc]"}, strings.Repeat(" ", 78) + "[a"},
	}
}
