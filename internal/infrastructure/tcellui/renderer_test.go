package tcellui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestRenderer_render_DrawsTargetsHUDAndStatusBar(t *testing.T) {
	// Given
	screen := newScreenFixture(80, 25)
	state := outbound.FrameViewState{
		Targets:   []outbound.TargetViewState{{X: 0, Y: 5, Tag: "[1234 victim]"}},
		HUD:       outbound.HUDViewState{Kills: 7},
		StatusBar: outbound.StatusViewState{Alive: 3, Speed: 1.0},
	}
	renderer := newRenderer(screen, outbound.ChromeSize{Top: 1, Bottom: 1}, state)

	// When
	renderer.render()

	// Then
	assert.Contains(t, screenRow(screen, 5), "[1234 victim]")
	assert.Contains(t, screenRow(screen, 0), "KILLS: 7")
	assert.Contains(t, screenRow(screen, 24), "Targets: 3")
}

func TestRenderer_render_ClearsPreviousFrame(t *testing.T) {
	// Given
	screen := newScreenFixture(80, 25)
	chrome := outbound.ChromeSize{Top: 1, Bottom: 1}
	previous := outbound.FrameViewState{
		Targets: []outbound.TargetViewState{{X: 0, Y: 5, Tag: "[1234 victim]"}},
	}
	first := newRenderer(screen, chrome, previous)
	first.render()
	second := newRenderer(screen, chrome, outbound.FrameViewState{})

	// When
	second.render()

	// Then
	assert.Empty(t, screenRow(screen, 5))
}
