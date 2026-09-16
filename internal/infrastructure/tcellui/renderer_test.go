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
	ui.Render(outbound.FrameState{HUD: outbound.HUDState{}})

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
	assert.NotContains(t, got, "Highscore", "row 0 should NOT contain Highscore on narrow terminal")
}

func TestDrawHUDWideTerminalDrawsAllThree(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewTUI(screen)
	require.NoError(t, ui.Init())
	defer ui.Cleanup()

	ui.Render(outbound.FrameState{HUD: outbound.HUDState{}})

	cells, w, _ := screen.GetContents()
	var row0 strings.Builder
	for x := range w {
		if r := cells[x].Runes; len(r) > 0 {
			row0.WriteRune(r[0])
		}
	}
	got := row0.String()

	for _, want := range []string{"FREED", "Highscore", "KILLS"} {
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
			ui.Render(outbound.FrameState{HUD: outbound.HUDState{Kills: 7}})

			assert.Equal(t, tc.want, rowContent(screen, 0))
		})
	}
}

func TestRenderMultiByteLabelColumnLayout(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	ui := tcellui.NewTUI(screen)
	require.NoError(t, ui.Init())
	defer ui.Cleanup()

	ui.Render(outbound.FrameState{
		Targets: []outbound.TargetViewState{
			{X: 0, Y: 2, Tag: "✦ KILLED ✦"},
		},
	})

	cells, w, _ := screen.GetContents()
	assert.Equal(t, ' ', cells[2*w+1].Runes[0], "col 1 should be space (rune after ✦) — byte-offset bug in render loop?")
	assert.Equal(t, 'K', cells[2*w+2].Runes[0], "col 2 should be 'K' — byte-offset bug in render loop?")
}

func TestRenderKillingIgnoresTagAndShowsAnimationFrame(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameState{
		Targets: []outbound.TargetViewState{{X: 0, Y: 5, Tag: "[1234 victim]", Killing: true, AnimationProgress: 0}},
	})

	got := rowContent(screen, 5)
	assert.NotContains(t, got, "victim", "Killing should ignore Tag and draw its own animation frame")
	assert.Contains(t, got, "💥", "the first kill-animation frame should render at progress 0")
}

func TestRenderKillingShowsTextFrameAtMidProgress(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameState{
		Targets: []outbound.TargetViewState{{X: 0, Y: 5, Killing: true, AnimationProgress: 0.3}},
	})

	got := rowContent(screen, 5)
	assert.Contains(t, got, "KILLED")
}

func TestRenderFleeingIgnoresTagAndShowsAnimationFrame(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameState{
		Targets: []outbound.TargetViewState{{X: 0, Y: 5, Tag: "[1234 victim]", Fleeing: true, AnimationProgress: 0}},
	})

	got := rowContent(screen, 5)
	assert.NotContains(t, got, "victim", "Fleeing should ignore Tag and draw its own animation frame")
	assert.Contains(t, got, "🏃", "the first flee-animation frame should render at progress 0")
}

func TestRenderWideRuneTagDoesNotDropCharacters(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameState{
		Targets: []outbound.TargetViewState{{X: 0, Y: 5, Tag: "[9 日本語]"}},
	})

	got := rowContent(screen, 5)
	for _, want := range []string{"[", "9", "日", "本", "語", "]"} {
		assert.Contains(t, got, want, "wide-rune tag should render %q without dropping characters", want)
	}
}

func TestRenderClipsTargetAtHUDRow(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameState{
		Targets: []outbound.TargetViewState{{X: 2, Y: 0, Tag: "[1234 victim]"}},
		HUD:     outbound.HUDState{Kills: 3},
	})

	got := rowContent(screen, 0)
	assert.NotContains(t, got, "victim", "target tag must not be drawn on the HUD row")
}

func TestDrawStatusBarNormal(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{Alive: 3, Speed: 2.0},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	for _, want := range []string{"Targets:", "Speed:", "Click to kill"} {
		assert.Contains(t, got, want, "status bar should contain %q", want)
	}
}

func TestDrawStatusBarConfirming(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{
			Confirming: &outbound.ConfirmViewState{PID: 42, Name: "myapp"},
		},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	for _, want := range []string{"42", "myapp", "(Y)es", "(N)o"} {
		assert.Contains(t, got, want, "confirm bar should contain %q", want)
	}
}

func TestDrawStatusBarConfirmingMultiByteName(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{
			Confirming: &outbound.ConfirmViewState{PID: 42, Name: "café-server"},
		},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.Contains(t, got, "café-server", "status bar should render the multi-byte name intact, with no byte-offset gap")
	assert.Contains(t, got, "(Y)es", "status bar should still contain the confirm options after a multi-byte name")
}

func TestDrawStatusBarConfirmingLongNameStillShowsAllOptions(t *testing.T) {
	for _, w := range []int{40, 44, 50} {
		t.Run(fmt.Sprintf("w=%d", w), func(t *testing.T) {
			ui, screen := newTUI(t)
			screen.SetSize(w, 25)
			ui.Render(outbound.FrameState{
				StatusBar: outbound.StatusState{
					Confirming: &outbound.ConfirmViewState{PID: 54321, Name: "com.apple.WebKit"},
				},
			})
			_, _, h := screen.GetContents()
			got := rowContent(screen, h-1)
			for _, want := range []string{"(Y)es", "(N)o", "(Q)uit"} {
				assert.Contains(t, got, want, "the confirm options must survive even with a long process name")
			}
		})
	}
}

func TestDrawStatusBarConfirmingLongNameIsTruncatedWithEllipsis(t *testing.T) {
	ui, screen := newTUI(t)
	screen.SetSize(44, 25)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{
			Confirming: &outbound.ConfirmViewState{PID: 54321, Name: "com.apple.WebKit"},
		},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.Contains(t, got, "…", "a truncated name should be signalled with an ellipsis")
	assert.NotContains(t, got, "com.apple.WebKit", "the full name should not fit at this width")
}

func TestDrawStatusBarConfirmingNameBudgetOfOneIsJustEllipsis(t *testing.T) {
	ui, screen := newTUI(t)
	screen.SetSize(38, 25)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{
			Confirming: &outbound.ConfirmViewState{PID: 54321, Name: "com.apple.WebKit"},
		},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.Contains(t, got, "[54321 …]", "with only one column of budget, the name should render as just the ellipsis")
}

func TestDrawStatusBarConfirmingNameOmittedWhenNoBudgetLeft(t *testing.T) {
	ui, screen := newTUI(t)
	screen.SetSize(30, 25)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{
			Confirming: &outbound.ConfirmViewState{PID: 54321, Name: "com.apple.WebKit"},
		},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.Contains(t, got, "[54321 ]", "with no budget left, the name should be dropped rather than partially shown")
	assert.NotContains(t, got, "com.apple.WebKit")
}

func TestDrawStatusBarWithTimeLimit(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{Alive: 1, Speed: 1.0, TimeLimit: 30, TimeLeft: 15},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.Contains(t, got, "Time:")
	assert.Contains(t, got, "15s")
}

func TestDrawStatusBarNoTimeLimit(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameState{
		StatusBar: outbound.StatusState{Alive: 1, Speed: 1.0, TimeLimit: 0},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.NotContains(t, got, "Time:")
}

// rowContent reads the visible characters on the given screen row.
func rowContent(screen tcell.SimulationScreen, row int) string {
	cells, w, _ := screen.GetContents()
	var sb strings.Builder
	for x := range w {
		if r := cells[row*w+x].Runes; len(r) > 0 {
			sb.WriteRune(r[0])
		}
	}
	return strings.TrimRight(sb.String(), " ")
}
