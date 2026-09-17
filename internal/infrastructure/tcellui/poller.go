package tcellui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/input"
)

// poller reads tcell events from a screen and translates them into game
// input events.
type poller struct {
	screen     tcell.Screen
	eventQueue chan input.Event
	// Reads as poller is done, writes as poller is stopping.
	done chan struct{}
}

// eventQueueCapacity is the buffer size of the translated input event
// channel, letting poll keep draining tcell events without blocking while
// the consumer is busy for up to this many events.
const eventQueueCapacity = 10

// newPoller returns a poller reading from screen. Call poll to start
// draining events; call stop to signal it to exit.
func newPoller(screen tcell.Screen) poller {
	return poller{
		screen:     screen,
		eventQueue: make(chan input.Event, eventQueueCapacity),
		done:       make(chan struct{}),
	}
}

func (p *poller) poll() {
	defer close(p.eventQueue)
	for {
		pollEvent := p.screen.PollEvent()
		if pollEvent == nil {
			return
		}
		inputEvent, ok := p.translateInputEvent(pollEvent)
		if !ok {
			continue
		}
		select {
		case p.eventQueue <- inputEvent:
		case <-p.done:
			return
		}
	}
}

// translateInputEvent converts a raw tcell event into a game input event.
// The second return value is false for events with no corresponding input
// event: an unrecognized key, a dropped mouse action, or a resize (handled
// here by re-syncing the screen, not by producing an event).
func (p *poller) translateInputEvent(pollEvent tcell.Event) (input.Event, bool) {
	switch pollEvent := pollEvent.(type) {
	case *tcell.EventMouse:
		return p.translateMouseEvent(pollEvent)
	case *tcell.EventKey:
		return translateKeyEvent(pollEvent)
	case *tcell.EventResize:
		p.screen.Sync()
	}
	return nil, false
}

// translateMouseEvent maps a left click outside the window chrome rows to
// a ClickEvent. Any other mouse activity — a different button, a chord
// (e.g. Button1+Button2 held together), or a click on a chrome row — is
// dropped.
func (p *poller) translateMouseEvent(mouseEvent *tcell.EventMouse) (input.Event, bool) {
	if mouseEvent.Buttons() != tcell.Button1 {
		return nil, false
	}
	x, y := mouseEvent.Position()
	_, height := p.screen.Size()
	if isChromeRow(y, height, chromeSize) {
		return nil, false
	}
	return input.ClickEvent{X: x, Y: y}, true
}

// stop signals poll to exit.
func (p *poller) stop() {
	close(p.done)
}

// events returns the channel of translated game input events.
func (p *poller) events() <-chan input.Event {
	return p.eventQueue
}

// keyBindings maps a rune to the game input event it produces.
type keyBindings map[rune]input.Event

var runeBindings = keyBindings{
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

// controlKeyBindings maps a non-rune control key to the game input event it
// produces.
var controlKeyBindings = map[tcell.Key]input.Event{
	tcell.KeyEscape: input.QuitEvent{},
	tcell.KeyCtrlC:  input.QuitEvent{},
	tcell.KeyCtrlZ:  input.QuitEvent{},
}

// translateKeyEvent maps a recognized key press to a game input event. A
// key matches by its rune or Key value alone; modifiers (e.g. Alt, Shift)
// are not considered.
func translateKeyEvent(keyEvent *tcell.EventKey) (input.Event, bool) {
	if inputEvent, ok := controlKeyBindings[keyEvent.Key()]; ok {
		return inputEvent, true
	}
	inputEvent, ok := runeBindings[keyEvent.Rune()]
	return inputEvent, ok
}
