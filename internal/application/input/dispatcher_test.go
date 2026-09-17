package input

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fixture"
)

func newDispatcher(s *game.Session) *Dispatcher {
	return NewDispatcher(game.NewInput(s))
}

// --- ClickEvent ---

func TestDispatcherClickReturnsHitTarget(t *testing.T) {
	s := fixture.Game(fixture.Processes(1), game.Config{Speed: 1.0})
	d := newDispatcher(s)
	tgt := s.Targets()[0]

	result := d.Dispatch(ClickEvent{X: int(math.Round(tgt.Position.X)), Y: int(math.Round(tgt.Position.Y))})

	assert.Equal(t, tgt, result)
}

func TestDispatcherClickMissReturnsNil(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 1.0})
	d := newDispatcher(s)

	result := d.Dispatch(ClickEvent{X: 0, Y: 0})

	assert.Nil(t, result)
}

// --- QuitEvent ---

func TestDispatcherQuitEventStopsGame(t *testing.T) {
	s := fixture.Game(nil, game.Config{})

	newDispatcher(s).Dispatch(QuitEvent{})

	assert.False(t, s.IsRunning())
}

// --- ConfirmEvent ---

func TestDispatcherConfirmEventAcceptReturnsConfirmedTarget(t *testing.T) {
	s := fixture.PendingConfirmGameSession(1)
	d := newDispatcher(s)
	tgt := s.Targets()[0]

	result := d.Dispatch(ConfirmEvent{Accept: true})

	assert.Equal(t, tgt, result)
	assert.Nil(t, s.PendingConfirm())
}

func TestDispatcherConfirmEventDeclineCancelsConfirmation(t *testing.T) {
	s := fixture.PendingConfirmGameSession(1)
	d := newDispatcher(s)

	d.Dispatch(ConfirmEvent{Accept: false})

	assert.Nil(t, s.PendingConfirm())
}

// --- SpeedEvent ---

func TestDispatcherSpeedEventFasterIncreasesSpeed(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 2.0})

	newDispatcher(s).Dispatch(SpeedEvent{Faster: true})

	assert.Equal(t, 2.5, s.Throttle().Speed())
}

func TestDispatcherSpeedEventSlowerDecreasesSpeed(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 2.0})

	newDispatcher(s).Dispatch(SpeedEvent{Faster: false})

	assert.Equal(t, 1.5, s.Throttle().Speed())
}
