package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestNewSession(t *testing.T) {
	processes := []process.Info{
		process.NewInfo(1, "a", 100, 0),
		process.NewInfo(2, "b", 200, 0),
	}

	s := NewSession(processes, Config{Confirm: true, Speed: 3.5, TimeLimit: 60})

	assert.True(t, s.cfg.Confirm)
	assert.Equal(t, 3.5, s.throttle.Speed())
	assert.Equal(t, 60, s.cfg.TimeLimit)
	assert.Equal(t, pending, s.currentState())
}

func TestStartTransitionsToRunning(t *testing.T) {
	processes := []process.Info{process.NewInfo(1, "a", 100, 0)}
	s := NewSession(processes, Config{})

	s.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))

	assert.True(t, s.IsRunning())
	assert.Len(t, s.roster.targets, 1)
}

func TestStopTransitionsToStopped(t *testing.T) {
	s := NewSession(nil, Config{})
	s.Start(movement.NewBounds(movement.WindowSize{Width: 0, Height: 0}, movement.ChromeSize{Top: 1, Bottom: 1}))

	s.Stop()

	assert.Equal(t, stopped, s.currentState())
}

func TestStartPanicsWhenRunning(t *testing.T) {
	s := NewSession(nil, Config{})
	s.Start(movement.NewBounds(movement.WindowSize{Width: 0, Height: 0}, movement.ChromeSize{Top: 1, Bottom: 1}))

	assert.Panics(t, func() {
		s.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	})
}

func TestStartPanicsWhenStopped(t *testing.T) {
	s := NewSession(nil, Config{})
	s.Start(movement.NewBounds(movement.WindowSize{Width: 0, Height: 0}, movement.ChromeSize{Top: 1, Bottom: 1}))
	s.Stop()

	assert.Panics(t, func() {
		s.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	})
}

func TestGameTimeLimit(t *testing.T) {
	s := NewSession(nil, Config{TimeLimit: 30})
	assert.Equal(t, 30, s.TimeLimit())
}

func TestGameTargetsEmptyBeforeStart(t *testing.T) {
	s := NewSession([]process.Info{process.NewInfo(1, "a", 0, 0)}, Config{})
	assert.Empty(t, s.Targets())
}

func TestGameTargetsPopulatedAfterStart(t *testing.T) {
	s := NewSession([]process.Info{process.NewInfo(1, "a", 0, 0)}, Config{})
	s.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	assert.Len(t, s.Targets(), 1)
}

func TestGamePendingConfirmNilWhenNoPending(t *testing.T) {
	s := NewSession(nil, Config{})
	assert.Nil(t, s.PendingConfirm())
}

func TestGamePendingConfirmReturnsPendingTarget(t *testing.T) {
	tgt := &Target{Info: process.NewInfo(42, "suspect", 0, 0)}
	s := &Session{confirm: confirmation{target: tgt, confirm: true}}
	assert.Equal(t, tgt, s.PendingConfirm())
}

func TestGameConfirmPendingFalseInitially(t *testing.T) {
	s := NewSession(nil, Config{})
	assert.False(t, s.confirm.Pending())
}

func TestGameThrottleMutationAffectsSpeed(t *testing.T) {
	s := NewSession(nil, Config{Speed: 2.0})
	s.Throttle().Increase()
	assert.Equal(t, 2.5, s.Throttle().Speed())
}

func TestGameAvailableTargetsExcludesDeadTargets(t *testing.T) {
	s := NewSession([]process.Info{process.NewInfo(1, "a", 0, 0)}, Config{Speed: 1.0})
	s.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	s.roster.targets[0].Kill()
	for range AnimationDuration {
		s.Update(movement.WindowSize{Width: 80, Height: 24})
	}

	targets, alive := s.AvailableTargets()

	assert.Empty(t, targets)
	assert.Equal(t, 0, alive)
}

func TestGameAvailableTargetsCountsAlive(t *testing.T) {
	processes := []process.Info{
		process.NewInfo(1, "a", 0, 0),
		process.NewInfo(2, "b", 0, 0),
	}
	s := NewSession(processes, Config{Speed: 1.0})
	s.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	s.roster.targets[0].Kill()

	targets, alive := s.AvailableTargets()

	require.Len(t, targets, 2) // killing + alive both available
	assert.Equal(t, 1, alive)
}

func TestUpdateStopsWhenTimeLimitExpired(t *testing.T) {
	clock := &fakeClock{t: time.Now()}
	s := &Session{timer: newTimer(1, clock.now)}
	s.Start(movement.NewBounds(movement.WindowSize{Width: 0, Height: 0}, movement.ChromeSize{Top: 1, Bottom: 1}))
	clock.advance(2 * time.Second)

	s.Update(movement.WindowSize{Width: 80, Height: 24})

	assert.False(t, s.IsRunning())
}

func TestUpdateStopsWhenAllTargetsDead(t *testing.T) {
	tgt := &Target{Info: process.NewInfo(1, "target", 0, 0), State: Dead}
	s := &Session{roster: roster{targets: []*Target{tgt}}, timer: newTimer(0, time.Now), throttle: movement.NewThrottle(movement.MinSpeed)}
	s.Start(movement.NewBounds(movement.WindowSize{Width: 0, Height: 0}, movement.ChromeSize{Top: 1, Bottom: 1}))

	s.Update(movement.WindowSize{Width: 80, Height: 24})

	assert.False(t, s.IsRunning())
}
