package tcellui_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestDrawStatusBarNormal(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameViewState{
		StatusBar: outbound.StatusViewState{Alive: 3, Speed: 2.0},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	for _, want := range []string{"Targets:", "Speed:", "Click to kill"} {
		assert.Contains(t, got, want, "status bar should contain %q", want)
	}
}

func TestDrawStatusBarConfirming(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameViewState{
		StatusBar: outbound.StatusViewState{
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
	ui.Render(outbound.FrameViewState{
		StatusBar: outbound.StatusViewState{
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
			ui.Render(outbound.FrameViewState{
				StatusBar: outbound.StatusViewState{
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
	ui.Render(outbound.FrameViewState{
		StatusBar: outbound.StatusViewState{
			Confirming: &outbound.ConfirmViewState{PID: 54321, Name: "com.apple.WebKit"},
		},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.Contains(t, got, "…", "a truncated name should be signaled with an ellipsis")
	assert.NotContains(t, got, "com.apple.WebKit", "the full name should not fit at this width")
}

func TestDrawStatusBarConfirmingNameBudgetOfOneIsJustEllipsis(t *testing.T) {
	ui, screen := newTUI(t)
	screen.SetSize(38, 25)
	ui.Render(outbound.FrameViewState{
		StatusBar: outbound.StatusViewState{
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
	ui.Render(outbound.FrameViewState{
		StatusBar: outbound.StatusViewState{
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
	timeLeft := 15
	ui.Render(outbound.FrameViewState{
		StatusBar: outbound.StatusViewState{Alive: 1, Speed: 1.0, TimeLeft: &timeLeft},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.Contains(t, got, "Time:")
	assert.Contains(t, got, "15s")
}

func TestDrawStatusBarNoTimeLimit(t *testing.T) {
	ui, screen := newTUI(t)
	ui.Render(outbound.FrameViewState{
		StatusBar: outbound.StatusViewState{Alive: 1, Speed: 1.0},
	})
	_, _, h := screen.GetContents()
	got := rowContent(screen, h-1)
	assert.NotContains(t, got, "Time:")
}
