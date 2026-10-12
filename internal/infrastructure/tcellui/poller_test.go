package tcellui

import (
	"testing"
	"testing/synctest"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestPoller_poll_ForwardsOnlyRecognizedEvents(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		screen := newScreenFixture(80, 25)
		defer screen.Fini()
		poller := newPoller(screen)
		go poller.poll()

		// When
		screen.InjectKey(tcell.KeyRune, 'z', tcell.ModNone)
		screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
		event := <-poller.Events()

		// Then
		assert.Equal(t, input.QuitEvent{}, event)
	})
}

func TestPoller_poll_ClosesQueueWhenScreenIsFinalized(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		screen := newScreenFixture(80, 25)
		poller := newPoller(screen)
		go poller.poll()

		// When
		screen.Fini()
		synctest.Wait()
		closed := testutil.IsClosed(poller.Events())

		// Then
		assert.True(t, closed)
	})
}

func TestPoller_stop_ClosesQueueWhileBlockedOnFullQueue(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		screen := newScreenFixture(80, 25)
		defer screen.Fini()
		poller := newPoller(screen)
		go poller.poll()
		for range eventQueueCapacity + 1 {
			screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
		}
		synctest.Wait()

		// When
		poller.stop()
		synctest.Wait()
		closed := testutil.IsClosed(poller.Events())

		// Then
		assert.True(t, closed)
	})
}
