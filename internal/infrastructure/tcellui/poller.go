package tcellui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
)

// eventQueueCapacity is the buffer size of the translated input event
// channel.
const eventQueueCapacity = 10

// Compile-time assertion that *poller satisfies outbound.InputEventProvider.
var _ outbound.InputEventProvider = (*poller)(nil)

// poller reads tcell events from a screen and translates them into input
// events.
type poller struct {
	screen     tcell.Screen
	eventQueue chan input.Event
	translator translator
	// done is closed by stop to tell poll to exit.
	done chan struct{}
}

// newPoller returns a poller reading from screen. Call poll to start
// draining events; call stop to signal it to exit.
func newPoller(screen tcell.Screen) poller {
	return poller{
		screen:     screen,
		eventQueue: make(chan input.Event, eventQueueCapacity),
		translator: newTranslator(screen),
		done:       make(chan struct{}),
	}
}

// poll queues the input events translated from screen's events until the
// screen is finalized or stop is called, then closes the event queue.
func (p *poller) poll() {
	defer close(p.eventQueue)
	for {
		pollEvent := p.screen.PollEvent()
		if pollEvent == nil {
			return
		}
		event, ok := p.translator.translate(pollEvent)
		if !ok {
			continue
		}
		select {
		case p.eventQueue <- event:
		case <-p.done:
			return
		}
	}
}

// stop signals poll to exit.
func (p *poller) stop() {
	close(p.done)
}

// Events returns the channel of translated input events.
func (p *poller) Events() <-chan input.Event {
	return p.eventQueue
}
