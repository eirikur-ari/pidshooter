package game_test

import (
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/testutil/fixture"
)

func TestInput_OnQuit_StopsGameSessionWhenThereIsNoPendingKillConfirmation(t *testing.T) {
	// Given
	gameSession := fixture.GameSession(nil, game.Config{})

	// When
	game.NewInput(gameSession).OnQuit()

	// Then
	assert.False(t, gameSession.IsRunning())
}

func TestInput_OnQuit_CancelsPendingKillConfirmation(t *testing.T) {
	// Given
	gameSession := fixture.PendingConfirmGameSession()
	input := game.NewInput(gameSession)

	// When
	input.OnQuit()

	// Then
	assert.True(t, gameSession.IsRunning(), "OnQuit should cancel confirm, not stop, when confirm is pending")
	assert.Nil(t, gameSession.PendingConfirm())
}

func TestInput_OnYes_AcceptsPendingKillConfirmationAndReturnsTargetToKill(t *testing.T) {
	// Given
	gameSession := fixture.PendingConfirmGameSession()
	input := game.NewInput(gameSession)

	// When
	result := input.OnYes()

	// Then
	require.NotNil(t, result)
	assert.Nil(t, gameSession.PendingConfirm())
}

func TestInput_OnNo_CancelsPendingKillConfirmation(t *testing.T) {
	// Given
	gameSession := fixture.PendingConfirmGameSession()
	input := game.NewInput(gameSession)

	// When
	input.OnNo()

	// Then
	assert.Nil(t, gameSession.PendingConfirm())
	assert.True(t, gameSession.IsRunning())
}

func TestInput_OnSpeedUp_IncreasesSpeed(t *testing.T) {
	// Given
	gameSession := fixture.GameSession(nil, game.Config{Speed: 2.0})

	// When
	game.NewInput(gameSession).OnSpeedUp()

	// Then
	assert.Equal(t, 2.5, gameSession.Throttle().Speed())
}

func TestInput_OnSpeedDown_DecreasesSpeed(t *testing.T) {
	// Given
	gameSession := fixture.GameSession(nil, game.Config{Speed: 2.0})

	// When
	game.NewInput(gameSession).OnSpeedDown()

	// Then
	assert.Equal(t, 1.5, gameSession.Throttle().Speed())
}

func TestInput_OnClickAt_ReturnsTargetOnHitWhenThereIsNoPendingKillConfirmation(t *testing.T) {
	// Given
	gameSession := fixture.GameSession(fixture.Processes(1), game.Config{Speed: 1.0})
	target := gameSession.Targets()[0]
	x, y := target.Motion.Position.Rounded()

	// When
	result := game.NewInput(gameSession).OnClickAt(x, y)

	// Then
	require.NotNil(t, result)
}

func TestInput_OnClickAt_ReturnsNilWhenHitAtIsAMiss(t *testing.T) {
	// Given
	gameSession := fixture.GameSession(nil, game.Config{Speed: 1.0})

	// When
	result := game.NewInput(gameSession).OnClickAt(0, 0)

	// Then
	assert.Nil(t, result)
}

func TestInput_OnClickAt_NoOpWhenAlreadyConfirming(t *testing.T) {
	// Given
	gameSession := fixture.PendingConfirmGameSession()
	input := game.NewInput(gameSession)
	target := gameSession.Targets()[0]

	// When
	result := input.OnClickAt(0, 0)

	// Then
	assert.Nil(t, result, "expected no result when already confirming")
	assert.Equal(t, target, gameSession.PendingConfirm(), "expected the original pending target to remain unchanged")
}
