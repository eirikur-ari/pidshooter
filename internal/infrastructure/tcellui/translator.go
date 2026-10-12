package tcellui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/input"
)

// runeBindings maps a rune to the input event it produces.
var runeBindings = map[rune]input.Event{
	'q': input.QuitEvent{},
	'Q': input.QuitEvent{},
	'y': input.ConfirmEvent{Accept: true},
	'Y': input.ConfirmEvent{Accept: true},
	'n': input.ConfirmEvent{Accept: false},
	'N': input.ConfirmEvent{Accept: false},
	'+': input.SpeedEvent{Faster: true},
	'=': input.SpeedEvent{Faster: true},
	'-': input.SpeedEvent{Faster: false},
	'_': input.SpeedEvent{Faster: false},
}

// controlKeyBindings maps a non-rune control key to the input event it
// produces.
var controlKeyBindings = map[tcell.Key]input.Event{
	tcell.KeyEscape: input.QuitEvent{},
	tcell.KeyCtrlC:  input.QuitEvent{},
	tcell.KeyCtrlZ:  input.QuitEvent{},
}

// translator translates a raw tcell.Event into the input event it
// represents.
type translator struct {
	screen      tcell.Screen
	prevButtons tcell.ButtonMask
}

// newTranslator returns a translator for the events of screen.
func newTranslator(screen tcell.Screen) translator {
	return translator{screen: screen}
}

// translate returns the input event rawEvent represents and reports whether
// it represents one. It reports false for an unrecognized key, a dropped
// mouse action, or a resize (handled here by re-syncing the screen, not by
// producing an event).
func (t *translator) translate(rawEvent tcell.Event) (input.Event, bool) {
	switch rawEvent := rawEvent.(type) {
	case *tcell.EventMouse:
		return t.translateMouseEvent(rawEvent)
	case *tcell.EventKey:
		return t.translateKeyEvent(rawEvent)
	case *tcell.EventResize:
		t.screen.Sync()
	}
	return nil, false
}

// translateMouseEvent translates the press that starts a left click to a
// ClickEvent. Any other mouse activity — a different button, more than one
// button held at once, or a motion sample or chord change while the left
// button is already held — is dropped.
func (t *translator) translateMouseEvent(mouseEvent *tcell.EventMouse) (input.Event, bool) {
	buttons := mouseEvent.Buttons()
	pressed := buttons == tcell.Button1 && t.prevButtons&tcell.Button1 == 0
	t.prevButtons = buttons
	if !pressed {
		return nil, false
	}
	x, y := mouseEvent.Position()
	return input.ClickEvent{X: x, Y: y}, true
}

// translateKeyEvent translates a recognized key press to an input event. A
// key matches by its rune or Key value alone; modifiers (e.g. Alt, Shift)
// are not considered.
func (t *translator) translateKeyEvent(keyEvent *tcell.EventKey) (input.Event, bool) {
	if event, ok := controlKeyBindings[keyEvent.Key()]; ok {
		return event, true
	}
	event, ok := runeBindings[keyEvent.Rune()]
	return event, ok
}
