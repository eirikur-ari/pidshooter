package input

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestClickEvent_dispatch_ReturnsHitTarget(t *testing.T) {
	// Given
	session := newGameSessionFixture([]process.Info{process.NewInfo(1, "a", 0, 0)}, game.Config{Speed: 1.0})
	target := session.Targets()[0]
	x, y := target.Motion.Position.Rounded()

	// When
	result := newDispatcherFixture(session).Dispatch(ClickEvent{X: x, Y: y})

	// Then
	assert.Equal(t, target, result)
}

func TestClickEvent_dispatch_ReturnsNilWhenClickMissesEveryTarget(t *testing.T) {
	// Given
	session := newGameSessionFixture(nil, game.Config{Speed: 1.0})

	// When
	result := newDispatcherFixture(session).Dispatch(ClickEvent{X: 0, Y: 0})

	// Then
	assert.Nil(t, result)
}

func TestQuitEvent_dispatch_StopsGameAndReturnsNil(t *testing.T) {
	// Given
	session := newGameSessionFixture(nil, game.Config{})

	// When
	result := newDispatcherFixture(session).Dispatch(QuitEvent{})

	// Then
	assert.False(t, session.IsRunning())
	assert.Nil(t, result)
}

func TestConfirmEvent_dispatch_ReturnsConfirmedTargetWhenAccepted(t *testing.T) {
	// Given
	session := pendingConfirmGameSessionFixture()
	target := session.Targets()[0]

	// When
	result := newDispatcherFixture(session).Dispatch(ConfirmEvent{Accept: true})

	// Then
	assert.Equal(t, target, result)
	assert.Nil(t, session.PendingConfirm())
}

func TestConfirmEvent_dispatch_CancelsConfirmationAndReturnsNilWhenDeclined(t *testing.T) {
	// Given
	session := pendingConfirmGameSessionFixture()

	// When
	result := newDispatcherFixture(session).Dispatch(ConfirmEvent{Accept: false})

	// Then
	assert.Nil(t, session.PendingConfirm())
	assert.Nil(t, result)
}

func TestSpeedEvent_dispatch_IncreasesSpeedAndReturnsNilWhenFaster(t *testing.T) {
	// Given
	session := newGameSessionFixture(nil, game.Config{Speed: 2.0})

	// When
	result := newDispatcherFixture(session).Dispatch(SpeedEvent{Faster: true})

	// Then
	assert.Equal(t, 2.5, session.Throttle().Speed())
	assert.Nil(t, result)
}

func TestSpeedEvent_dispatch_DecreasesSpeedAndReturnsNilWhenSlower(t *testing.T) {
	// Given
	session := newGameSessionFixture(nil, game.Config{Speed: 2.0})

	// When
	result := newDispatcherFixture(session).Dispatch(SpeedEvent{Faster: false})

	// Then
	assert.Equal(t, 1.5, session.Throttle().Speed())
	assert.Nil(t, result)
}
