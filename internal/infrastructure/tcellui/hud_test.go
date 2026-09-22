package tcellui_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
)

func TestDrawHUDNarrowTerminalSuppressesCenter(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewTUI(screen)
	require.NoError(t, ui.Init())
	defer ui.Cleanup()

	screen.SetSize(30, 25)
	ui.Render(outbound.FrameViewState{HUD: outbound.HUDViewState{}})

	cells, w, _ := screen.GetContents()
	var row0 strings.Builder
	for x := range w {
		if r := cells[x].Runes; len(r) > 0 {
			row0.WriteRune(r[0])
		}
	}
	got := row0.String()

	assert.Contains(t, got, "FREED")
	assert.Contains(t, got, "KILLS")
	assert.NotContains(t, got, "HighScore", "row 0 should NOT contain HighScore on narrow terminal")
}

func TestDrawHUDWideTerminalDrawsAllThree(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewTUI(screen)
	require.NoError(t, ui.Init())
	defer ui.Cleanup()

	ui.Render(outbound.FrameViewState{HUD: outbound.HUDViewState{}})

	cells, w, _ := screen.GetContents()
	var row0 strings.Builder
	for x := range w {
		if r := cells[x].Runes; len(r) > 0 {
			row0.WriteRune(r[0])
		}
	}
	got := row0.String()

	for _, want := range []string{"FREED", "HighScore", "KILLS"} {
		assert.Contains(t, got, want, "row 0 should contain %q on wide terminal", want)
	}
}

// TestDrawHUDFreedLabelBudgetedAgainstKillsColumn is a regression test for
// docs/tcellui-findings.md Minor Finding 7: FREED must be truncated to the
// KILLS column's budget rather than relying on KILLS's later draw call to
// overwrite whatever runs past it.
func TestDrawHUDFreedLabelBudgetedAgainstKillsColumn(t *testing.T) {
	cases := []struct {
		width int
		want  string
	}{
		{10, " KILLS: 7"},
		{14, " FRE KILLS: 7"},
		{18, " FREED:  KILLS: 7"},
		{24, " FREED: 0 B    KILLS: 7"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("w=%d", tc.width), func(t *testing.T) {
			ui, screen := newTUI(t)
			screen.SetSize(tc.width, 25)
			ui.Render(outbound.FrameViewState{HUD: outbound.HUDViewState{Kills: 7}})

			assert.Equal(t, tc.want, rowContent(screen, 0))
		})
	}
}
