package tcellui_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
)

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

func TestRenderDrawsMultipleTargets(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameState{
		Targets: []outbound.TargetViewState{
			{X: 0, Y: 5, Tag: "[1111 first]"},
			{X: 0, Y: 8, Tag: "[2222 second]"},
		},
	})

	assert.Contains(t, rowContent(screen, 5), "first", "the first target should be drawn")
	assert.Contains(t, rowContent(screen, 8), "second", "the second target should be drawn, not just the first")
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
