package tcellui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/input"
)

func TestTranslateKeyEvent(t *testing.T) {
	tests := newTranslateKeyEventTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tr := translator{}
			ok := tr.translateKeyEvent(tcell.NewEventKey(test.key, test.r, tcell.ModNone))
			assert.Equal(t, test.wantOK, ok)
			if test.wantOK {
				assert.Equal(t, test.want, tr.event)
			}
		})
	}
}

func TestTranslateMouseEvent(t *testing.T) {
	tests := newTranslateMouseEventTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tr := translator{}
			ok := tr.translateMouseEvent(tcell.NewEventMouse(test.x, test.y, test.button, tcell.ModNone))
			assert.Equal(t, test.wantOK, ok)
			if test.wantOK {
				assert.Equal(t, test.want, tr.event)
			}
		})
	}
}

func newTranslateKeyEventTestCases() []struct {
	name   string
	key    tcell.Key
	r      rune
	wantOK bool
	want   input.EventDispatcher
} {
	tests := []struct {
		name   string
		key    tcell.Key
		r      rune
		wantOK bool
		want   input.EventDispatcher
	}{
		{"escape quits", tcell.KeyEscape, 0, true, input.QuitEvent{}},
		{"ctrl+c quits", tcell.KeyCtrlC, 0, true, input.QuitEvent{}},
		{"ctrl+z quits", tcell.KeyCtrlZ, 0, true, input.QuitEvent{}},
		{"q quits", tcell.KeyRune, 'q', true, input.QuitEvent{}},
		{"uppercase Q quits", tcell.KeyRune, 'Q', true, input.QuitEvent{}},
		{"y confirms accept", tcell.KeyRune, 'y', true, input.ConfirmEvent{Accept: true}},
		{"uppercase Y confirms accept", tcell.KeyRune, 'Y', true, input.ConfirmEvent{Accept: true}},
		{"n confirms decline", tcell.KeyRune, 'n', true, input.ConfirmEvent{Accept: false}},
		{"uppercase N confirms decline", tcell.KeyRune, 'N', true, input.ConfirmEvent{Accept: false}},
		{"+ speeds up", tcell.KeyRune, '+', true, input.SpeedEvent{Faster: true}},
		{"= speeds up", tcell.KeyRune, '=', true, input.SpeedEvent{Faster: true}},
		{"- speeds down", tcell.KeyRune, '-', true, input.SpeedEvent{Faster: false}},
		{"_ speeds down", tcell.KeyRune, '_', true, input.SpeedEvent{Faster: false}},
		{"unrecognized rune dropped", tcell.KeyRune, 'z', false, nil},
	}
	return tests
}

func newTranslateMouseEventTestCases() []struct {
	name   string
	x, y   int
	button tcell.ButtonMask
	wantOK bool
	want   input.EventDispatcher
} {
	tests := []struct {
		name   string
		x, y   int
		button tcell.ButtonMask
		wantOK bool
		want   input.EventDispatcher
	}{
		{"button1 emits click", 5, 10, tcell.Button1, true, input.ClickEvent{X: 5, Y: 10}},
		{"non-button1 dropped", 5, 10, tcell.Button2, false, nil},
		{"click on row 0 emits click", 5, 0, tcell.Button1, true, input.ClickEvent{X: 5, Y: 0}},
	}
	return tests
}
