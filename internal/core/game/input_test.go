package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestInput_OnQuit_StopsGameSessionWhenThereIsNoPendingKillConfirmation(t *testing.T) {
	// Given
	gameSession := newSessionFixture(nil, Config{})

	// When
	NewInput(gameSession).OnQuit()
	isRunning := gameSession.IsRunning()

	// Then
	assert.False(t, isRunning)
}

func TestInput_OnQuit_CancelsPendingKillConfirmation(t *testing.T) {
	// Given
	gameSession := pendingConfirmSessionFixture()
	input := NewInput(gameSession)

	// When
	input.OnQuit()
	isRunning := gameSession.IsRunning()
	pendingConfirm := gameSession.PendingConfirm()

	// Then
	assert.True(t, isRunning, "OnQuit should cancel confirm, not stop, when confirm is pending")
	assert.Nil(t, pendingConfirm)
}

func TestInput_OnYes_AcceptsPendingKillConfirmationAndReturnsTargetToKill(t *testing.T) {
	// Given
	gameSession := pendingConfirmSessionFixture()
	input := NewInput(gameSession)

	// When
	result := input.OnYes()
	pendingConfirm := gameSession.PendingConfirm()

	// Then
	require.NotNil(t, result)
	assert.Nil(t, pendingConfirm)
}

func TestInput_OnNo_CancelsPendingKillConfirmation(t *testing.T) {
	// Given
	gameSession := pendingConfirmSessionFixture()
	input := NewInput(gameSession)

	// When
	input.OnNo()
	pendingConfirm := gameSession.PendingConfirm()
	isRunning := gameSession.IsRunning()

	// Then
	assert.Nil(t, pendingConfirm)
	assert.True(t, isRunning)
}

func TestInput_OnSpeedUp_IncreasesSpeed(t *testing.T) {
	// Given
	gameSession := newSessionFixture(nil, Config{Speed: 2.0})

	// When
	NewInput(gameSession).OnSpeedUp()
	speed := gameSession.Throttle().Speed()

	// Then
	assert.Equal(t, 2.5, speed)
}

func TestInput_OnSpeedDown_DecreasesSpeed(t *testing.T) {
	// Given
	gameSession := newSessionFixture(nil, Config{Speed: 2.0})

	// When
	NewInput(gameSession).OnSpeedDown()
	speed := gameSession.Throttle().Speed()

	// Then
	assert.Equal(t, 1.5, speed)
}

func TestInput_OnClickAt_ReturnsTargetOnHitWhenThereIsNoPendingKillConfirmation(t *testing.T) {
	// Given
	gameSession := newSessionFixture([]process.Info{process.NewInfo(1, "a", 0, 0)}, Config{Speed: 1.0})
	target := gameSession.Targets()[0]
	x, y := target.Motion.Position.Rounded()

	// When
	result := NewInput(gameSession).OnClickAt(x, y)

	// Then
	require.NotNil(t, result)
}

func TestInput_OnClickAt_ReturnsNilWhenHitAtIsAMiss(t *testing.T) {
	// Given
	gameSession := newSessionFixture(nil, Config{Speed: 1.0})

	// When
	result := NewInput(gameSession).OnClickAt(0, 0)

	// Then
	assert.Nil(t, result)
}

func TestInput_OnClickAt_NoOpWhenAlreadyConfirming(t *testing.T) {
	// Given
	gameSession := pendingConfirmSessionFixture()
	input := NewInput(gameSession)
	target := gameSession.Targets()[0]

	// When
	result := input.OnClickAt(0, 0)
	pendingConfirm := gameSession.PendingConfirm()

	// Then
	assert.Nil(t, result, "expected no result when already confirming")
	assert.Equal(t, target, pendingConfirm, "expected the original pending target to remain unchanged")
}
