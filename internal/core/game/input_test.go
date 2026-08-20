package game_test

import (
	"math"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil/fixture"
)

func newInput(s *game.Session) *game.Input {
	return game.NewInput(s)
}

// --- OnQuit ---

func TestInputOnQuitStopsGame(t *testing.T) {
	s := fixture.Game(nil, game.Config{})

	newInput(s).OnQuit()

	assert.False(t, s.IsRunning())
}

func TestInputOnQuitCancelsConfirmWhenPending(t *testing.T) {
	s := fixture.PendingConfirmGameSession(1)
	in := newInput(s)

	in.OnQuit()

	assert.True(t, s.IsRunning(), "OnQuit should cancel confirm, not stop, when confirm is pending")
	assert.Nil(t, s.PendingConfirm())
}

// --- OnYes ---

func TestInputOnYesReturnsConfirmedTarget(t *testing.T) {
	s := fixture.PendingConfirmGameSession(1)
	in := newInput(s)

	result := in.OnYes()

	require.NotNil(t, result)
	assert.Nil(t, s.PendingConfirm())
}

// --- OnNo ---

func TestInputOnNoCancelsPending(t *testing.T) {
	s := fixture.PendingConfirmGameSession(1)
	in := newInput(s)

	in.OnNo()

	assert.Nil(t, s.PendingConfirm())
	assert.True(t, s.IsRunning())
}

// --- OnSpeedUp / OnSpeedDown ---

func TestInputOnSpeedUpIncreasesSpeed(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 2.0})

	newInput(s).OnSpeedUp()

	assert.Equal(t, 2.5, s.Throttle().Speed())
}

func TestInputOnSpeedDownDecreasesSpeed(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 2.0})

	newInput(s).OnSpeedDown()

	assert.Equal(t, 1.5, s.Throttle().Speed())
}

// --- OnClickAt ---

func TestInputOnClickAtReturnsTarget(t *testing.T) {
	s := fixture.Game(fixture.Processes(1), game.Config{Speed: 1.0})
	tgt := s.Targets()[0]

	result := newInput(s).OnClickAt(int(math.Round(tgt.Position.X)), int(math.Round(tgt.Position.Y)))

	require.NotNil(t, result)
}

func TestInputOnClickAtMissReturnsNil(t *testing.T) {
	s := fixture.Game(nil, game.Config{Speed: 1.0})

	result := newInput(s).OnClickAt(0, 0)

	assert.Nil(t, result)
}

func TestInputOnClickAtNoOpWhenAlreadyConfirming(t *testing.T) {
	s := fixture.PendingConfirmGameSession(2)
	in := newInput(s)
	pending := s.PendingConfirm()

	second := s.Targets()[1]
	result := in.OnClickAt(int(math.Round(second.Position.X)), int(math.Round(second.Position.Y)))

	assert.Nil(t, result, "expected no result when already confirming")
	assert.Equal(t, pending, s.PendingConfirm(), "expected pending confirmation unchanged")
}
