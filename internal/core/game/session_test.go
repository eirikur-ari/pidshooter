package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestSession_Start_TransitionsFromPendingToRunning(t *testing.T) {
	// Given
	processes := []process.Info{newInfoFixture()}
	session := NewSession(processes, Config{})

	// When
	pendingState := session.currentState()
	session.Start(newBoundsFixture())
	runningState := session.currentState()
	result := session.IsRunning()

	// Then
	assert.Equal(t, pending, pendingState)
	assert.Equal(t, running, runningState)
	assert.True(t, result)
}

func TestSession_Start_PanicsWhenAlreadyRunning(t *testing.T) {
	// Given
	session := NewSession(nil, Config{})

	// When
	session.Start(newBoundsFixture())

	// Then
	assert.Panics(t, func() {
		session.Start(newBoundsFixture())
	})
}

func TestSession_Start_PanicsWhenAlreadyStopped(t *testing.T) {
	// Given
	session := NewSession(nil, Config{})

	// When
	session.Start(newBoundsFixture())
	session.Stop()

	// Then
	assert.Panics(t, func() {
		session.Start(newBoundsFixture())
	})
}

func TestSession_Stop_TransitionsFromRunningToStopped(t *testing.T) {
	// Given
	session := NewSession(nil, Config{})
	bounds := newBoundsFixture()

	// When
	session.Start(bounds)
	runningState := session.currentState()
	session.Stop()
	stoppedState := session.currentState()
	result := session.IsRunning()

	// Then
	assert.Equal(t, running, runningState)
	assert.Equal(t, stopped, stoppedState)
	assert.False(t, result)
}

func TestSession_Update_AdvancesTargetWhileRunning(t *testing.T) {
	// Given
	info := newInfoFixture()
	info.PID = 1
	info.Name = "x"
	session := NewSession([]process.Info{info}, Config{Speed: 1.0})
	bounds := newBoundsFixture()

	// When
	session.Start(bounds)
	target := session.Targets()[0]
	target.Motion = movement.Motion{Position: movement.Vector{X: 35, Y: 10}, Velocity: movement.Vector{X: 2.0, Y: 0}}

	session.Update(movement.WindowSize{Width: 40, Height: 24})
	isRunning := session.IsRunning()

	// Then
	assert.True(t, isRunning)
	assert.Equal(t, 35.0, target.Motion.Position.X, "the narrower window passed to Update, not Start, must constrain movement")
	assert.Less(t, target.Motion.Velocity.X, 0.0, "hitting the new right wall should bounce velocity")
}

func TestSession_Update_StopsWhenTimeLimitExpired(t *testing.T) {
	// Given
	clock := &FakeClock{T: time.Now()}
	session := &Session{timer: newTimer(1, clock.Now)}
	bounds := newBoundsFixture()

	// When
	session.Start(bounds)
	clock.Advance(2 * time.Second)
	session.Update(windowSizeFixture())
	isRunning := session.IsRunning()

	// Then
	assert.False(t, isRunning)
}

func TestSession_Update_StopsWhenAllTargetsAreDead(t *testing.T) {
	// Given
	target := &Target{Info: newInfoFixture(), State: Dead}
	clock := &FakeClock{T: time.Now()}
	session := &Session{roster: roster{targets: []*Target{target}}, timer: newTimer(0, clock.Now), throttle: movement.NewThrottle(movement.MinSpeed)}
	bounds := newBoundsFixture()

	// When
	session.Start(bounds)
	session.Update(windowSizeFixture())

	// Then
	assert.False(t, session.IsRunning())
}

func TestSession_PendingConfirm_ReturnsNilWhenNoPendingTarget(t *testing.T) {
	// Given
	session := NewSession(nil, Config{})

	// When
	result := session.PendingConfirm()

	// Then
	assert.Nil(t, result)
}

func TestSession_PendingConfirm_ReturnsPendingTarget(t *testing.T) {
	// Given
	target := &Target{Info: newInfoFixture()}
	session := &Session{confirm: confirmation{target: target, confirm: true}}

	// When
	result := session.PendingConfirm()

	// Then
	assert.Equal(t, target, result)
}

func TestSession_Targets_EmptyByDefault(t *testing.T) {
	// Given
	session := NewSession([]process.Info{newInfoFixture()}, Config{})

	// When
	result := session.Targets()

	// Then
	assert.Empty(t, result)
}

func TestSession_Targets_PopulatedAfterStart(t *testing.T) {
	// Given
	session := NewSession([]process.Info{newInfoFixture()}, Config{})
	bounds := newBoundsFixture()

	// When
	session.Start(bounds)
	result := session.Targets()

	// Then
	assert.Len(t, result, 1)
}

func TestSession_Throttle_ReflectsConfiguredSpeed(t *testing.T) {
	// Given
	session := NewSession(nil, Config{Speed: 2.0})

	// When
	session.Throttle().Increase()

	// Then
	assert.Equal(t, 2.5, session.Throttle().Speed())
}
