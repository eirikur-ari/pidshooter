package tcellui

import (
	"errors"
	"testing"
	"testing/synctest"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
)

func TestNewTUI_PanicsOnNilScreen(t *testing.T) {
	assert.Panics(t, func() { NewTUI(nil) })
}

func TestTUI_Init_StartsForwardingInputEvents(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		screen := newUninitializedScreenFixture()
		ui := NewTUI(screen)
		defer ui.Cleanup()

		// When
		require.NoError(t, ui.Init())
		screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
		event := <-ui.InputEvents().Events()

		// Then
		assert.Equal(t, input.QuitEvent{}, event)
	})
}

func TestTUI_Init_WrapsScreenInitFailure(t *testing.T) {
	// Given
	cause := errors.New("no terminal")
	screen := &MockScreen{}
	screen.On("Init").Return(cause)
	ui := NewTUI(screen)

	// When
	err := ui.Init()

	// Then
	require.Error(t, err)
	assert.ErrorIs(t, err, cause)
	assert.EqualError(t, err, "failed to initialize screen: no terminal")
}

func TestTUI_Init_RejectsSecondCallWithoutSecondPoller(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		ui := NewTUI(newUninitializedScreenFixture())
		require.NoError(t, ui.Init())

		// When
		err := ui.Init()
		ui.Cleanup()
		synctest.Wait()

		// Then
		assert.EqualError(t, err, "already initialized")
		assert.True(t, isClosed(ui.InputEvents().Events()))
	})
}

func TestTUI_Cleanup_ClosesInputEvents(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		ui := NewTUI(newUninitializedScreenFixture())
		require.NoError(t, ui.Init())

		// When
		ui.Cleanup()
		synctest.Wait()
		closed := isClosed(ui.InputEvents().Events())

		// Then
		assert.True(t, closed)
	})
}

func TestTUI_Cleanup_CanBeCalledRepeatedly(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		ui := NewTUI(newUninitializedScreenFixture())
		require.NoError(t, ui.Init())
		ui.Cleanup()

		// When
		repeated := func() { ui.Cleanup() }

		// Then
		assert.NotPanics(t, repeated)
	})
}

func TestTUI_Cleanup_LeavesScreenAloneWhenNeverInitialized(t *testing.T) {
	// Given
	screen := &MockScreen{}
	ui := NewTUI(screen)

	// When
	ui.Cleanup()

	// Then
	screen.AssertNotCalled(t, "Fini")
}

func TestTUI_WindowSize_ReturnsScreenDimensions(t *testing.T) {
	// Given
	screen := newScreenFixture(42, 17)
	ui := NewTUI(screen)

	// When
	window := ui.WindowSize()

	// Then
	assert.Equal(t, outbound.WindowSize{Width: 42, Height: 17}, window)
}

func TestTUI_ChromeSize_ReservesOneRowTopAndBottom(t *testing.T) {
	// Given
	ui := NewTUI(newScreenFixture(80, 25))

	// When
	chrome := ui.ChromeSize()

	// Then
	assert.Equal(t, outbound.ChromeSize{Top: 1, Bottom: 1}, chrome)
}

func TestTUI_Render_KeepsTargetsOffTheChromeRows(t *testing.T) {
	// Given
	screen := newScreenFixture(80, 25)
	ui := NewTUI(screen)
	state := outbound.FrameViewState{
		Targets: []outbound.TargetViewState{{X: 20, Y: 0, Tag: "[1234 victim]"}},
	}

	// When
	ui.Render(state)

	// Then
	assert.NotContains(t, screenRow(screen, 0), "victim")
}
