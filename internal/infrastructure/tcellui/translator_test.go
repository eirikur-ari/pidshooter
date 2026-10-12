package tcellui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/input"
)

func TestTranslator_translate_RoutesRawEventsByType(t *testing.T) {
	tests := newRoutedEventTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			translator := translator{}

			// When
			event, ok := translator.translate(test.rawEvent)

			// Then
			assert.True(t, ok)
			assert.Equal(t, test.expected, event)
		})
	}
}

func TestTranslator_translate_ResyncsScreenOnResizeWithoutEvent(t *testing.T) {
	// Given
	screen := &MockScreen{}
	screen.On("Sync").Return()
	translator := newTranslator(screen)

	// When
	event, ok := translator.translate(tcell.NewEventResize(100, 40))

	// Then
	screen.AssertNumberOfCalls(t, "Sync", 1)
	assert.False(t, ok)
	assert.Nil(t, event)
}

func TestTranslator_translate_IgnoresUnknownEventTypes(t *testing.T) {
	// Given
	translator := translator{}

	// When
	event, ok := translator.translate(tcell.NewEventInterrupt(nil))

	// Then
	assert.False(t, ok)
	assert.Nil(t, event)
}

func TestTranslator_translateMouseEvent_ReportsClickAtPressPosition(t *testing.T) {
	tests := newClickPositionTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			translator := translator{}
			mouseEvent := tcell.NewEventMouse(test.x, test.y, tcell.Button1, tcell.ModNone)

			// When
			event, ok := translator.translateMouseEvent(mouseEvent)

			// Then
			assert.True(t, ok)
			assert.Equal(t, test.expected, event)
		})
	}
}

func TestTranslator_translateMouseEvent_IgnoresOtherButtons(t *testing.T) {
	tests := newIgnoredButtonsTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			translator := translator{}
			mouseEvent := tcell.NewEventMouse(5, 10, test.buttons, tcell.ModNone)

			// When
			event, ok := translator.translateMouseEvent(mouseEvent)

			// Then
			assert.False(t, ok)
			assert.Nil(t, event)
		})
	}
}

func TestTranslator_translateMouseEvent_ClicksOncePerLeftPress(t *testing.T) {
	tests := newPressSequenceTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			translator := translator{}

			// When
			var clicked []bool
			for _, buttons := range test.sequence {
				_, ok := translator.translateMouseEvent(tcell.NewEventMouse(1, 1, buttons, tcell.ModNone))
				clicked = append(clicked, ok)
			}

			// Then
			assert.Equal(t, test.expected, clicked)
		})
	}
}

func TestTranslator_translateKeyEvent_MapsRecognizedKeys(t *testing.T) {
	tests := newRecognizedKeyTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			translator := translator{}
			keyEvent := tcell.NewEventKey(test.key, test.r, test.modifiers)

			// When
			event, ok := translator.translateKeyEvent(keyEvent)

			// Then
			assert.True(t, ok)
			assert.Equal(t, test.expected, event)
		})
	}
}

func TestTranslator_translateKeyEvent_IgnoresUnrecognizedKeys(t *testing.T) {
	tests := newUnrecognizedKeyTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			translator := translator{}
			keyEvent := tcell.NewEventKey(test.key, test.r, tcell.ModNone)

			// When
			event, ok := translator.translateKeyEvent(keyEvent)

			// Then
			assert.False(t, ok)
			assert.Nil(t, event)
		})
	}
}

func newRoutedEventTestCase() []struct {
	name     string
	rawEvent tcell.Event
	expected input.Event
} {
	return []struct {
		name     string
		rawEvent tcell.Event
		expected input.Event
	}{
		{"key event", tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone), input.QuitEvent{}},
		{"mouse event", tcell.NewEventMouse(3, 4, tcell.Button1, tcell.ModNone), input.ClickEvent{X: 3, Y: 4}},
	}
}

func newClickPositionTestCase() []struct {
	name     string
	x, y     int
	expected input.Event
} {
	return []struct {
		name     string
		x, y     int
		expected input.Event
	}{
		{"press inside the window", 5, 10, input.ClickEvent{X: 5, Y: 10}},
		{"press on the top-left cell", 0, 0, input.ClickEvent{X: 0, Y: 0}},
	}
}

func newIgnoredButtonsTestCase() []struct {
	name    string
	buttons tcell.ButtonMask
} {
	return []struct {
		name    string
		buttons tcell.ButtonMask
	}{
		{"middle button", tcell.Button2},
		{"right button", tcell.Button3},
		{"left and right together", tcell.Button1 | tcell.Button2},
		{"no button", tcell.ButtonNone},
		{"wheel", tcell.WheelUp},
	}
}

func newPressSequenceTestCase() []struct {
	name     string
	sequence []tcell.ButtonMask
	expected []bool
} {
	return []struct {
		name     string
		sequence []tcell.ButtonMask
		expected []bool
	}{
		{
			"motion while the button is held is not another click",
			[]tcell.ButtonMask{tcell.Button1, tcell.Button1},
			[]bool{true, false},
		},
		{
			"pressing again after release is another click",
			[]tcell.ButtonMask{tcell.Button1, tcell.ButtonNone, tcell.Button1},
			[]bool{true, false, true},
		},
		{
			"tapping a second button while the left is held is not another click",
			[]tcell.ButtonMask{tcell.Button1, tcell.Button1 | tcell.Button2, tcell.Button1},
			[]bool{true, false, false},
		},
		{
			"releasing the other button of a chord is not a click",
			[]tcell.ButtonMask{tcell.Button1 | tcell.Button2, tcell.Button1},
			[]bool{false, false},
		},
	}
}

func newRecognizedKeyTestCase() []struct {
	name      string
	key       tcell.Key
	r         rune
	modifiers tcell.ModMask
	expected  input.Event
} {
	return []struct {
		name      string
		key       tcell.Key
		r         rune
		modifiers tcell.ModMask
		expected  input.Event
	}{
		{"escape quits", tcell.KeyEscape, 0, tcell.ModNone, input.QuitEvent{}},
		{"ctrl+c quits", tcell.KeyCtrlC, 0, tcell.ModNone, input.QuitEvent{}},
		{"ctrl+z quits", tcell.KeyCtrlZ, 0, tcell.ModNone, input.QuitEvent{}},
		{"q quits", tcell.KeyRune, 'q', tcell.ModNone, input.QuitEvent{}},
		{"uppercase Q quits", tcell.KeyRune, 'Q', tcell.ModNone, input.QuitEvent{}},
		{"y confirms accept", tcell.KeyRune, 'y', tcell.ModNone, input.ConfirmEvent{Accept: true}},
		{"uppercase Y confirms accept", tcell.KeyRune, 'Y', tcell.ModNone, input.ConfirmEvent{Accept: true}},
		{"n confirms decline", tcell.KeyRune, 'n', tcell.ModNone, input.ConfirmEvent{Accept: false}},
		{"uppercase N confirms decline", tcell.KeyRune, 'N', tcell.ModNone, input.ConfirmEvent{Accept: false}},
		{"+ speeds up", tcell.KeyRune, '+', tcell.ModNone, input.SpeedEvent{Faster: true}},
		{"= speeds up", tcell.KeyRune, '=', tcell.ModNone, input.SpeedEvent{Faster: true}},
		{"- speeds down", tcell.KeyRune, '-', tcell.ModNone, input.SpeedEvent{Faster: false}},
		{"_ speeds down", tcell.KeyRune, '_', tcell.ModNone, input.SpeedEvent{Faster: false}},
		{"modifiers on a rune key are ignored", tcell.KeyRune, 'q', tcell.ModShift, input.QuitEvent{}},
		{"modifiers on a control key are ignored", tcell.KeyEscape, 0, tcell.ModAlt, input.QuitEvent{}},
	}
}

func newUnrecognizedKeyTestCase() []struct {
	name string
	key  tcell.Key
	r    rune
} {
	return []struct {
		name string
		key  tcell.Key
		r    rune
	}{
		{"unbound rune", tcell.KeyRune, 'z'},
		{"unbound control key", tcell.KeyEnter, 0},
		{"function key", tcell.KeyF1, 0},
	}
}
