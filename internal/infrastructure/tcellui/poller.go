package tcellui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/input"
)

// eventQueueCapacity is the buffer size of the translated input event
// channel.
const eventQueueCapacity = 10

// poller reads tcell events from a screen and translates them into input
// events.
type poller struct {
	screen     tcell.Screen
	eventQueue chan input.EventDispatcher
	translator translator
	// Reads as poller is done, writes as poller is stopping.
	done chan struct{}
}

// newPoller returns a poller reading from screen. Call poll to start
// draining events; call stop to signal it to exit.
func newPoller(screen tcell.Screen) poller {
	return poller{
		screen:     screen,
		eventQueue: make(chan input.EventDispatcher, eventQueueCapacity),
		translator: newTranslator(screen),
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
		if !p.translator.translateEvent(pollEvent) {
			continue
		}
		select {
		case p.eventQueue <- p.translator.event:
		case <-p.done:
			return
		}
	}
}

// stop signals poll to exit.
func (p *poller) stop() {
	close(p.done)
}

// events returns the channel of translated input events.
func (p *poller) events() <-chan input.EventDispatcher {
	return p.eventQueue
}
