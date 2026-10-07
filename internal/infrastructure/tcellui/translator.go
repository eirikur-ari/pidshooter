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
	event       input.Event
	prevButtons tcell.ButtonMask
}

// newTranslator returns a translator reading window state from screen.
func newTranslator(screen tcell.Screen) translator {
	return translator{screen: screen}
}

// translateEvent sets event to the input event rawEvent represents and reports
// whether one was produced. event is left unchanged when it reports false:
// an unrecognized key, a dropped mouse action, or a resize (handled here by
// re-syncing the screen, not by producing an event).
func (t *translator) translateEvent(rawEvent tcell.Event) bool {
	switch rawEvent := rawEvent.(type) {
	case *tcell.EventMouse:
		return t.translateMouseEvent(rawEvent)
	case *tcell.EventKey:
		return t.translateKeyEvent(rawEvent)
	case *tcell.EventResize:
		t.screen.Sync()
	}
	return false
}

// translateMouseEvent translates the press that starts a left click to a
// ClickEvent. Any other mouse activity — a different button, more than one
// button held at once, or a motion sample while the button is already held
// — is dropped.
func (t *translator) translateMouseEvent(mouseEvent *tcell.EventMouse) bool {
	buttons := mouseEvent.Buttons()
	pressed := buttons == tcell.Button1 && t.prevButtons != tcell.Button1
	t.prevButtons = buttons
	if !pressed {
		return false
	}
	x, y := mouseEvent.Position()
	t.event = input.ClickEvent{X: x, Y: y}
	return true
}

// translateKeyEvent translates a recognized key press to an input event. A
// key matches by its rune or Key value alone; modifiers (e.g. Alt, Shift)
// are not considered.
func (t *translator) translateKeyEvent(keyEvent *tcell.EventKey) bool {
	if event, ok := controlKeyBindings[keyEvent.Key()]; ok {
		t.event = event
		return true
	}
	event, ok := runeBindings[keyEvent.Rune()]
	if !ok {
		return false
	}
	t.event = event
	return true
}
