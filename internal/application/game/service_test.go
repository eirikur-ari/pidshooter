package game

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// --- Play ---

func TestServicePlayInvalidSpeedReturnsClassifiedError(t *testing.T) {
	svc := NewService(nil, nil, nil)

	_, err := svc.Play(PlayRequest{Speed: 99}, nil, 0)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.ErrorContains(t, err, "speed must be between")
}

func TestServicePlayInvalidTimeLimitReturnsClassifiedError(t *testing.T) {
	svc := NewService(nil, nil, nil)

	_, err := svc.Play(PlayRequest{Speed: 2.0, TimeLimit: -1}, nil, 0)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.ErrorContains(t, err, "time must be")
}

func TestServicePlayNoProcessesReturnsClassifiedError(t *testing.T) {
	svc := NewService(nil, nil, nil)

	_, err := svc.Play(PlayRequest{Speed: 2.0}, nil, 0)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessNotFound, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.Equal(t, "no processes found", err.Error(), "an empty wrapper Message should not change the displayed text")
	require.Error(t, appErr.Unwrap(), "the underlying cause should still be reachable, not discarded")
}

func TestServicePlayRendererInitFailureReturnsClassifiedError(t *testing.T) {
	processes := []process.Info{process.NewInfo(100, "target", 4096, 0)}
	renderer := &fake.Renderer{InitErr: errors.New("terminal not available")}
	svc := NewService(nil, renderer, fake.NewInputEventProvider())

	_, err := svc.Play(PlayRequest{Speed: 2.0}, processes, 0)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeGameFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.ErrorContains(t, err, "terminal not available")
}

func TestServicePlaySuccessReturnsPlayResult(t *testing.T) {
	processes := []process.Info{process.NewInfo(100, "target", 4096, 0)}
	events := fake.NewInputEventProvider()
	events.Ch <- input.QuitEvent{}
	svc := NewService(nil, &fake.Renderer{}, events)

	result, err := svc.Play(PlayRequest{Speed: 2.0}, processes, 0)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Duration, 0.0)
	assert.Equal(t, 2.0, result.LowestSpeed, "quitting before any speed change should leave the starting speed unchanged")
	assert.Equal(t, 0, result.Kills)
	assert.Equal(t, int64(0), result.FreedMem)
	assert.Empty(t, result.KillFailures)
}

// --- drainEventQueue ---

func TestDrainEventQueueReturnsErrorWhenEventChannelCloses(t *testing.T) {
	events := fake.NewInputEventProvider()
	close(events.Ch)

	svc := NewService(nil, &fake.Renderer{}, events)
	killSignals := make(chan killSignal, 1)
	done := make(chan struct{})
	defer close(done)

	err := svc.drainEventQueue(nil, killSignals, done, &sync.WaitGroup{})

	assert.EqualError(t, err, "input event channel closed")
}

// --- applyKillSignals ---

func TestServiceApplyKillsCompletesPendingKill(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputEventProvider())
	kills := make(chan killSignal, 1)

	target := game.NewTarget(info, movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	kills <- killSignal{target: target}

	tracker := newKillTracker(0)
	svc.applyKillSignals(tracker, kills)

	assert.Equal(t, game.Killing, target.State)
	assert.Equal(t, 1, tracker.score.kills)
}

func TestServiceApplyKillsReapsAlreadyKilledTarget(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputEventProvider())
	kills := make(chan killSignal, 1)

	target := game.NewTarget(info, movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	kills <- killSignal{target: target, shouldReap: true}

	tracker := newKillTracker(0)
	svc.applyKillSignals(tracker, kills)

	assert.Equal(t, game.Fleeing, target.State, "an already-gone target should flee rather than die outright")
	assert.Equal(t, 0, tracker.score.kills, "reaping an already-gone target should not award a kill")
	require.Len(t, tracker.duds, 1)
	assert.Equal(t, "target", tracker.duds[0].Target)
	assert.Equal(t, 100, tracker.duds[0].PID)
}

func TestServiceApplyKillsEmptyChannelNoOps(t *testing.T) {
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputEventProvider())
	kills := make(chan killSignal, 1)

	tracker := newKillTracker(0)
	svc.applyKillSignals(tracker, kills) // must not block

	assert.Equal(t, 0, tracker.score.kills)
}

func TestServiceApplyKillsRecordsFailureWithoutMutatingTarget(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputEventProvider())
	kills := make(chan killSignal, 1)

	target := game.NewTarget(info, movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	kills <- killSignal{target: target, err: errors.New("operation not permitted")}

	tracker := newKillTracker(0)
	svc.applyKillSignals(tracker, kills)

	assert.Equal(t, game.Alive, target.State, "a kill failure must leave the target alive")
	assert.Equal(t, 0, tracker.score.kills)
	require.Len(t, tracker.failure.failures, 1)
	failure := tracker.failure.failures[0]
	assert.Equal(t, "target", failure.Target)
	assert.Equal(t, 100, failure.PID)
	assert.EqualError(t, failure.Err, "operation not permitted")
}

func TestServiceApplyKillsDeduplicatesRepeatedFailuresForSamePID(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputEventProvider())
	kills := make(chan killSignal, 2)

	target := game.NewTarget(info, movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	kills <- killSignal{target: target, err: errors.New("operation not permitted")}
	kills <- killSignal{target: target, err: errors.New("operation not permitted")}

	tracker := newKillTracker(0)
	svc.applyKillSignals(tracker, kills)

	assert.Len(t, tracker.failure.failures, 1, "a repeat failure for the same PID should not duplicate the entry")
}

// TestKillOrReapReportsFailureWithoutReaping covers the two cases that used
// to be silently dropped: a protected-PID refusal and a kill syscall failure
// (both shouldReap=false, err!=nil). The target must stay unreaped and the
// failure must be reported on killSignals rather than swallowed.
func TestKillOrReapReportsFailureWithoutReaping(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	target := game.NewTarget(info, movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	killer := fake.ProcessKiller(func(int, string, bool) (bool, error) {
		return false, errors.New("refusing to kill PID 100")
	})
	svc := NewService(killer, &fake.Renderer{}, fake.NewInputEventProvider())
	killSignals := make(chan killSignal, 1)
	done := make(chan struct{})
	defer close(done)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	svc.killOrReap(target, killSignals, done, wg)

	sig := <-killSignals
	assert.False(t, sig.shouldReap)
	assert.ErrorContains(t, sig.err, "refusing to kill PID 100")
}

// --- killOrReap ---

// TestKillOrReapReapWithoutErrorStaysSilent verifies the existing silent-reap
// path (process already gone / PID recycled — shouldReap=true, err set) does
// not start reporting a failure now that killOrReap also surfaces them: the
// signal must carry shouldReap and no err, exactly as before this change.
func TestKillOrReapReapWithoutErrorStaysSilent(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	target := game.NewTarget(info, movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	killer := fake.ProcessKiller(func(int, string, bool) (bool, error) {
		return true, errors.New("could not verify PID 100: process not found")
	})
	svc := NewService(killer, &fake.Renderer{}, fake.NewInputEventProvider())
	killSignals := make(chan killSignal, 1)
	done := make(chan struct{})
	defer close(done)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	svc.killOrReap(target, killSignals, done, wg)

	sig := <-killSignals
	assert.True(t, sig.shouldReap)
	assert.NoError(t, sig.err, "a reap outcome must not be reported as a kill failure")
}

func TestRegisterTermSignalWatcherTermSignalClosesOnStop(t *testing.T) {
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputEventProvider())

	termSignal, stop := svc.registerTermSignalWatcher()
	stop()

	select {
	case <-termSignal:
	case <-time.After(time.Second):
		t.Fatal("termSignal was not closed after stop()")
	}
}

// --- registerTermSignalWatcher ---
