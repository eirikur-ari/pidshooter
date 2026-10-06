package game

import (
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestService_Play_RejectsInvalidRequest(t *testing.T) {
	tests := newInvalidPlayRequestTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			var appErr *apperror.Error
			service := NewService(nil, nil, nil)

			// When
			_, err := service.Play(test.request, nil, 0)

			// Then
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, apperror.CodeInvalidConfig, appErr.Code)
			assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
		})
	}
}

func TestService_Play_RejectsEmptyProcessList(t *testing.T) {
	tests := newEmptyProcessListTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			service := NewService(nil, nil, nil)
			var appErr *apperror.Error

			// When
			_, err := service.Play(PlayRequest{Speed: 2.0}, test.processes, 0)

			// Then
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, apperror.CodeProcessNotFound, appErr.Code)
			assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
			assert.Equal(t, "no processes found", err.Error())
			require.Error(t, appErr.Unwrap())
		})
	}
}

func TestService_Play_FailsWhenRendererInitFails(t *testing.T) {
	// Given
	renderer := &testutil.FakeRenderer{InitErr: errors.New("display not available")}
	service := NewService(nil, renderer, testutil.NewFakeInputEventProvider())
	var appErr *apperror.Error

	// When
	_, err := service.Play(PlayRequest{Speed: 2.0}, []process.Info{newInfoFixture()}, 0)

	// Then
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeGameFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.ErrorContains(t, err, "renderer initialization failed: display not available")
}

func TestService_Play_FailsAndCleansUpWhenEventChannelCloses(t *testing.T) {
	// Given
	events := testutil.NewFakeInputEventProvider()
	close(events.Ch)
	renderer := &testutil.FakeRenderer{}
	service := NewService(nil, renderer, events)
	var appErr *apperror.Error

	// When
	_, err := service.Play(PlayRequest{Speed: 2.0}, []process.Info{newInfoFixture()}, 0)

	// Then
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeGameFailed, appErr.Code)
	assert.Equal(t, 1, renderer.CleanupCalls)
}

func TestService_Play_ReturnsResultWhenPlayerQuits(t *testing.T) {
	// Given
	events := testutil.NewFakeInputEventProvider()
	events.Ch <- input.QuitEvent{}
	service := NewService(nil, &testutil.FakeRenderer{}, events)

	// When
	result, err := service.Play(PlayRequest{Speed: 2.0}, []process.Info{newInfoFixture()}, 0)

	// Then
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Duration, 0.0)
	assert.Equal(t, 2.0, result.LowestSpeed)
	assert.Equal(t, 0, result.Kills)
	assert.Equal(t, int64(0), result.FreedMem)
	assert.Empty(t, result.KillFailures)
	assert.Empty(t, result.Duds)
}

func TestService_drainEventQueue_ReturnsErrorWhenEventChannelCloses(t *testing.T) {
	// Given
	events := testutil.NewFakeInputEventProvider()
	close(events.Ch)
	service := NewService(nil, nil, events)
	done := make(chan struct{})
	defer close(done)

	// When
	err := service.drainEventQueue(nil, make(chan killSignal, 1), done, &sync.WaitGroup{})

	// Then
	assert.EqualError(t, err, "input event channel closed")
}

func TestService_drainEventQueue_IgnoresSecondClickOnTargetWithKillInFlight(t *testing.T) {
	// Given
	session, x, y := newClickableSessionFixture()
	events := testutil.NewFakeInputEventProvider()
	events.Ch <- input.ClickEvent{X: x, Y: y}
	events.Ch <- input.ClickEvent{X: x, Y: y}
	killer := &MockProcessKiller{}
	killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	service := NewService(killer, nil, events)
	dispatcher := input.NewDispatcher(game.NewInput(session))
	signals := make(chan killSignal, 10)
	done := make(chan struct{})
	defer close(done)
	var waitGroup sync.WaitGroup

	// When
	err := service.drainEventQueue(dispatcher, signals, done, &waitGroup)
	waitGroup.Wait()

	// Then
	require.NoError(t, err)
	killer.AssertNumberOfCalls(t, "Kill", 1)
	require.Len(t, signals, 1)
	assert.Equal(t, session.Targets()[0], (<-signals).target)
}

func TestService_frameLoop_AppliesKillToTrackerOnceItsAnimationFinishes(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		session, x, y := newClickableSessionFixture()
		events := testutil.NewFakeInputEventProvider()
		events.Ch <- input.ClickEvent{X: x, Y: y}
		killer := &MockProcessKiller{}
		killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
		service := NewService(killer, &testutil.FakeRenderer{}, events)
		dispatcher := input.NewDispatcher(game.NewInput(session))
		tracker := newKillTracker(0)

		// When
		_, err := service.frameLoop(session, tracker, dispatcher, nil)

		// Then
		require.NoError(t, err)
		assert.Equal(t, 1, tracker.score.kills)
		assert.Equal(t, newInfoFixture().Rss, tracker.score.freedMem)
		assert.Empty(t, tracker.failure.failures)
	})
}

func TestService_frameLoop_WaitsForKillInFlightWhenSessionStops(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		const killDelay = 300 * time.Millisecond
		session, x, y := newClickableSessionFixture()
		events := testutil.NewFakeInputEventProvider()
		events.Ch <- input.ClickEvent{X: x, Y: y}
		events.Ch <- input.QuitEvent{}
		killer := &MockProcessKiller{}
		killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(false, nil).After(killDelay)
		service := NewService(killer, &testutil.FakeRenderer{}, events)
		dispatcher := input.NewDispatcher(game.NewInput(session))
		tracker := newKillTracker(0)
		start := time.Now()

		// When
		endTime, err := service.frameLoop(session, tracker, dispatcher, nil)

		// Then
		require.NoError(t, err)
		assert.Equal(t, killDelay, time.Since(start), "frameLoop must wait for the in-flight kill before returning")
		assert.True(t, endTime.Equal(start), "the end time is taken before waiting for in-flight kills")
		assert.Equal(t, 1, tracker.score.kills, "a kill that lands after the session stops is still counted")
	})
}

func TestService_frameLoop_StopsSessionWhenTermSignalFires(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		session := newSessionFixture(game.Config{Speed: 1.0, TimeLimit: 60})
		service := NewService(nil, &testutil.FakeRenderer{}, testutil.NewFakeInputEventProvider())
		dispatcher := input.NewDispatcher(game.NewInput(session))
		termSignal := make(chan struct{}, 1)
		termSignal <- struct{}{}
		start := time.Now()

		// When
		_, err := service.frameLoop(session, newKillTracker(0), dispatcher, termSignal)

		// Then
		require.NoError(t, err)
		assert.False(t, session.IsRunning())
		assert.Equal(t, time.Duration(0), time.Since(start), "the session must stop at the signal, not run to its time limit")
	})
}

func TestService_awaitOutstandingKills_GivesUpOnceGracePeriodElapses(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		const gracePeriod = 20 * time.Millisecond
		service := &Service{killGracePeriod: gracePeriod}
		var waitGroup sync.WaitGroup
		waitGroup.Add(1)
		defer waitGroup.Done()
		start := time.Now()

		// When
		service.awaitOutstandingKills(&waitGroup, newKillTracker(0), make(chan killSignal, 1))

		// Then
		assert.Equal(t, gracePeriod, time.Since(start))
	})
}

func TestService_applyKillSignals_KillsTargetWhoseKillSucceeded(t *testing.T) {
	// Given
	service := NewService(nil, nil, nil)
	target := newTargetFixture()
	signals := make(chan killSignal, 1)
	signals <- killSignal{target: target}
	tracker := newKillTracker(0)

	// When
	service.applyKillSignals(tracker, signals)

	// Then
	assert.Equal(t, game.Killing, target.State)
	assert.Equal(t, 1, tracker.score.kills)
	assert.Equal(t, int64(4096), tracker.score.freedMem)
}

func TestService_applyKillSignals_ReapsTargetWhoseProcessWasAlreadyGone(t *testing.T) {
	// Given
	service := NewService(nil, nil, nil)
	target := newTargetFixture()
	signals := make(chan killSignal, 1)
	signals <- killSignal{target: target, shouldReap: true}
	tracker := newKillTracker(0)

	// When
	service.applyKillSignals(tracker, signals)

	// Then
	assert.Equal(t, game.Fleeing, target.State)
	assert.Equal(t, 0, tracker.score.kills)
	assert.Equal(t, []KillDud{{Name: "target", PID: 100}}, tracker.duds)
	assert.Len(t, tracker.duds, 1)
}

func TestService_applyKillSignals_IgnoresReapOfTargetThatIsNotAlive(t *testing.T) {
	// Given
	service := NewService(nil, nil, nil)
	target := newTargetFixture()
	target.Kill()
	signals := make(chan killSignal, 1)
	signals <- killSignal{target: target, shouldReap: true}
	tracker := newKillTracker(0)

	// When
	service.applyKillSignals(tracker, signals)

	// Then
	assert.Equal(t, game.Killing, target.State)
	assert.Empty(t, tracker.duds)
}

func TestService_applyKillSignals_RecordsFailureAndLeavesTargetAlive(t *testing.T) {
	// Given
	service := NewService(nil, nil, nil)
	target := newTargetFixture()
	cause := errors.New("operation not permitted")
	signals := make(chan killSignal, 1)
	signals <- killSignal{target: target, err: cause}
	tracker := newKillTracker(0)

	// When
	service.applyKillSignals(tracker, signals)

	// Then
	assert.Equal(t, game.Alive, target.State)
	assert.Equal(t, 0, tracker.score.kills)
	assert.Equal(t, []KillFailure{{Name: "target", PID: 100, Err: cause}}, tracker.failure.failures)
	assert.Len(t, tracker.failure.failures, 1)
}

func TestService_applyKillSignals_RecordsRepeatedFailuresOfSamePIDOnce(t *testing.T) {
	// Given
	service := NewService(nil, nil, nil)
	target := newTargetFixture()
	signals := make(chan killSignal, 2)
	signals <- killSignal{target: target, err: errors.New("operation not permitted")}
	signals <- killSignal{target: target, err: errors.New("operation not permitted")}
	tracker := newKillTracker(0)

	// When
	service.applyKillSignals(tracker, signals)

	// Then
	assert.Len(t, tracker.failure.failures, 1)
}

func TestService_applyKillSignals_ReturnsWithoutBlockingWhenNoSignalIsBuffered(t *testing.T) {
	// Given
	service := NewService(nil, nil, nil)
	tracker := newKillTracker(0)

	// When
	service.applyKillSignals(tracker, make(chan killSignal, 1))

	// Then
	assert.Equal(t, 0, tracker.score.kills)
}

func TestService_killOrReap_ReportsKillerOutcome(t *testing.T) {
	tests := newKillOrReapTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			killer := &MockProcessKiller{}
			killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(test.shouldReap, test.err)
			service := NewService(killer, nil, nil)
			target := newTargetFixture()
			signals := make(chan killSignal, 1)
			done := make(chan struct{})
			defer close(done)
			waitGroup := &sync.WaitGroup{}
			waitGroup.Add(1)
			expected := test.expected
			expected.target = target

			// When
			service.killOrReap(target, signals, done, waitGroup)

			// Then
			assert.Equal(t, expected, <-signals)
		})
	}
}

func TestService_killOrReap_PassesTargetDetailsToKiller(t *testing.T) {
	tests := newKillArgumentsTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			killer := &MockProcessKiller{}
			killer.On("Kill", test.expected.pid, test.expected.name, test.expected.protected).Return(false, nil)
			service := NewService(killer, nil, nil)
			done := make(chan struct{})
			defer close(done)
			waitGroup := &sync.WaitGroup{}
			waitGroup.Add(1)

			// When
			service.killOrReap(test.target, make(chan killSignal, 1), done, waitGroup)

			// Then
			killer.AssertExpectations(t)
		})
	}
}

func TestService_killOrReap_DropsOutcomeWhenDoneIsClosed(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		killer := &MockProcessKiller{}
		killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
		service := NewService(killer, nil, nil)
		signals := make(chan killSignal)
		done := make(chan struct{})
		close(done)
		var waitGroup sync.WaitGroup
		waitGroup.Add(1)

		// When
		service.killOrReap(newTargetFixture(), signals, done, &waitGroup)
		waitGroup.Wait()

		// Then
		killer.AssertNumberOfCalls(t, "Kill", 1)
	})
}

func TestService_registerTermSignalWatcher_ClosesTermSignalWhenStopped(t *testing.T) {
	// Given
	service := NewService(nil, nil, nil)
	termSignal, stop := service.registerTermSignalWatcher()

	// When
	stop()

	// Then
	select {
	case <-termSignal:
	case <-time.After(time.Second):
		t.Fatal("termSignal was not closed after stop()")
	}
}

func newInvalidPlayRequestTestCase() []struct {
	name    string
	request PlayRequest
} {
	return []struct {
		name    string
		request PlayRequest
	}{
		{"speed out of range", PlayRequest{Speed: 99}},
		{"negative time limit", PlayRequest{Speed: 2.0, TimeLimit: -1}},
	}
}

func newKillOrReapTestCase() []struct {
	name       string
	shouldReap bool
	err        error
	expected   killSignal
} {
	failure := errors.New("refusing to kill PID 100")

	return []struct {
		name       string
		shouldReap bool
		err        error
		expected   killSignal
	}{
		{"kill succeeded", false, nil, killSignal{}},
		{"kill failed", false, failure, killSignal{err: failure}},
		{"process already gone drops the verification error", true, errors.New("could not verify PID 100: process not found"), killSignal{shouldReap: true}},
	}
}

func newEmptyProcessListTestCase() []struct {
	name      string
	processes []process.Info
} {
	return []struct {
		name      string
		processes []process.Info
	}{
		{"nil list", nil},
		{"zero-length list", []process.Info{}},
	}
}

func newKillArgumentsTestCase() []struct {
	name     string
	target   *game.Target
	expected struct {
		pid       int
		name      string
		protected bool
	}
} {
	type arguments = struct {
		pid       int
		name      string
		protected bool
	}

	return []struct {
		name     string
		target   *game.Target
		expected arguments
	}{
		{"unprotected process", newTargetFixtureFor(100, "target"), arguments{pid: 100, name: "target", protected: false}},
		{"protected process", newTargetFixtureFor(1, "init"), arguments{pid: 1, name: "init", protected: true}},
	}
}
