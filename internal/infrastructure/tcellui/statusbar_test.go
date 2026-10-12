package tcellui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestStatusBar_draw_ShowsPlayStatus(t *testing.T) {
	tests := newPlayStatusTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			screen := newScreenFixture(80, 25)
			statusBar := statusBar{state: test.state, window: outbound.WindowSize{Width: 80, Height: 25}}

			// When
			statusBar.draw(screen)

			// Then
			assert.Equal(t, test.expected, screenRow(screen, 24))
		})
	}
}

func TestStatusBar_draw_ShowsConfirmPromptWithinWindowWidth(t *testing.T) {
	tests := newConfirmPromptTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			screen := newScreenFixture(test.width, 25)
			statusBar := statusBar{
				state: outbound.StatusViewState{
					Confirming: &outbound.ConfirmViewState{PID: 54321, Name: test.processName},
				},
				window: outbound.WindowSize{Width: test.width, Height: 25},
			}

			// When
			statusBar.draw(screen)

			// Then
			assert.Equal(t, test.expected, screenRow(screen, 24))
		})
	}
}

func TestStatusBar_draw_ReplacesPreviousRowContent(t *testing.T) {
	// Given
	screen := newScreenFixture(80, 25)
	screen.PutStrStyled(0, 24, strings.Repeat("X", 80), tcell.StyleDefault)
	statusBar := statusBar{
		state:  outbound.StatusViewState{Alive: 3, Speed: 2.0},
		window: outbound.WindowSize{Width: 80, Height: 25},
	}

	// When
	statusBar.draw(screen)

	// Then
	assert.Equal(t, " Targets: 3 | Speed: 2.0x | Click to kill | +/- speed | 'q' quit", screenRow(screen, 24))
}

func newPlayStatusTestCase() []struct {
	name     string
	state    outbound.StatusViewState
	expected string
} {
	timeLeft := 15
	return []struct {
		name     string
		state    outbound.StatusViewState
		expected string
	}{
		{
			"alive count and speed",
			outbound.StatusViewState{Alive: 3, Speed: 2.0},
			" Targets: 3 | Speed: 2.0x | Click to kill | +/- speed | 'q' quit",
		},
		{
			"time left is shown when the session has a time limit",
			outbound.StatusViewState{Alive: 1, Speed: 1.0, TimeLeft: &timeLeft},
			" Targets: 1 | Speed: 1.0x | Time: 15s | Click to kill | +/- speed | 'q' quit",
		},
	}
}

func newConfirmPromptTestCase() []struct {
	name        string
	processName string
	width       int
	expected    string
} {
	return []struct {
		name        string
		processName string
		width       int
		expected    string
	}{
		{"name that fits is shown whole", "com.apple.WebKit", 80, " Kill [54321 com.apple.WebKit]? (Y)es / (N)o / (Q)uit"},
		{"accented name stays intact", "café-server", 80, " Kill [54321 café-server]? (Y)es / (N)o / (Q)uit"},
		{"name too long is cut with an ellipsis", "com.apple.WebKit", 44, " Kill [54321 com.ap…]? (Y)es / (N)o / (Q)uit"},
		{"one column of budget is just the ellipsis", "com.apple.WebKit", 38, " Kill [54321 …]? (Y)es / (N)o / (Q)uit"},
		{"no budget drops the name", "com.apple.WebKit", 37, " Kill [54321 ]? (Y)es / (N)o / (Q)uit"},
		{"wide name is cut by columns", "日本語アプリ", 44, " Kill [54321 日本語…]? (Y)es / (N)o / (Q)uit"},
		{"wide name never splits a character", "日本語アプリ", 43, " Kill [54321 日本…]? (Y)es / (N)o / (Q)uit"},
		{"window narrower than the prompt clips the options", "com.apple.WebKit", 30, " Kill [54321 ]? (Y)es / (N)o /"},
	}
}
