package game

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestNewPlaySession_RejectsMissingSessionOrTracker(t *testing.T) {
	tests := newMissingCollaboratorTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			play, err := newPlaySession(nil, nil, nil, defaultKillGracePeriod, test.session, test.tracker)

			// Then
			assert.Nil(t, play)
			assert.EqualError(t, err, test.expected)
		})
	}
}

func TestPlaySession_result_MapsSessionAndTrackerState(t *testing.T) {
	// Given
	play := newPlaySessionFixture(t, nil, nil, nil, game.Config{Speed: 2.0})
	play.tracker.recordKill(2048)
	play.tracker.recordFailure(newTargetFixtureFor(200, "stubborn"), assert.AnError)
	play.tracker.recordDud(newTargetFixtureFor(300, "gone"))
	endTime := play.session.StartTime().Add(1500 * time.Millisecond)

	// When
	actual := play.result(endTime)

	// Then
	assert.Equal(t, PlayResult{
		Duration:     1.5,
		LowestSpeed:  2.0,
		Kills:        1,
		FreedMem:     2048,
		KillFailures: []KillFailure{{Name: "stubborn", PID: 200, Err: assert.AnError}},
		Duds:         []KillDud{{Name: "gone", PID: 300}},
	}, actual)
}

func TestPlaySession_drainEventQueue_ReturnsErrorWhenEventChannelCloses(t *testing.T) {
	// Given
	events := testutil.NewFakeInputEventProvider()
	close(events.Ch)
	play := newPlaySessionFixture(t, nil, nil, events, game.Config{Speed: 1.0})

	// When
	err := play.drainEventQueue()

	// Then
	assert.EqualError(t, err, "input event channel closed")
}

func TestPlaySession_drainEventQueue_IgnoresSecondClickOnTargetWithKillInFlight(t *testing.T) {
	// Given
	events := testutil.NewFakeInputEventProvider()
	killer := &MockProcessKiller{}
	killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	play := newPlaySessionFixture(t, killer, nil, events, game.Config{Speed: 1.0})
	x, y := play.session.Targets()[0].Motion.Position.Rounded()
	events.Ch <- input.ClickEvent{X: x, Y: y}
	events.Ch <- input.ClickEvent{X: x, Y: y}

	// When
	err := play.drainEventQueue()
	play.kills.awaitRemaining()

	// Then
	require.NoError(t, err)
	killer.AssertNumberOfCalls(t, "Kill", 1)
	assert.Equal(t, 1, play.tracker.score.kills)
}

func TestPlaySession_frameLoop_AppliesKillToTrackerOnceItsAnimationFinishes(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		events := testutil.NewFakeInputEventProvider()
		killer := &MockProcessKiller{}
		killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
		play := newPlaySessionFixture(t, killer, &testutil.FakeRenderer{}, events, game.Config{Speed: 1.0})
		x, y := play.session.Targets()[0].Motion.Position.Rounded()
		events.Ch <- input.ClickEvent{X: x, Y: y}

		// When
		_, err := play.frameLoop(nil)

		// Then
		require.NoError(t, err)
		assert.Equal(t, 1, play.tracker.score.kills)
		assert.Equal(t, newInfoFixture().Rss, play.tracker.score.freedMem)
		assert.Empty(t, play.tracker.failure.failures)
	})
}

func TestPlaySession_frameLoop_WaitsForKillInFlightWhenSessionStops(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		const killDelay = 300 * time.Millisecond
		events := testutil.NewFakeInputEventProvider()
		killer := &MockProcessKiller{}
		killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(false, nil).After(killDelay)
		play := newPlaySessionFixture(t, killer, &testutil.FakeRenderer{}, events, game.Config{Speed: 1.0})
		x, y := play.session.Targets()[0].Motion.Position.Rounded()
		events.Ch <- input.ClickEvent{X: x, Y: y}
		events.Ch <- input.QuitEvent{}
		start := time.Now()

		// When
		endTime, err := play.frameLoop(nil)

		// Then
		require.NoError(t, err)
		assert.Equal(t, killDelay, time.Since(start), "frameLoop must wait for the in-flight kill before returning")
		assert.True(t, endTime.Equal(start), "the end time is taken before waiting for in-flight kills")
		assert.Equal(t, 1, play.tracker.score.kills, "a kill that lands after the session stops is still counted")
	})
}

func TestPlaySession_frameLoop_StopsSessionWhenTermSignalFires(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		play := newPlaySessionFixture(t, nil, &testutil.FakeRenderer{}, testutil.NewFakeInputEventProvider(), game.Config{Speed: 1.0, TimeLimit: 60})
		termSignal := make(chan struct{}, 1)
		termSignal <- struct{}{}
		start := time.Now()

		// When
		_, err := play.frameLoop(termSignal)

		// Then
		require.NoError(t, err)
		assert.False(t, play.session.IsRunning())
		assert.Equal(t, time.Duration(0), time.Since(start), "the session must stop at the signal, not run to its time limit")
	})
}

func TestRegisterTermSignalWatcher_ClosesTermSignalWhenStopped(t *testing.T) {
	// Given
	termSignal, stop := registerTermSignalWatcher()

	// When
	stop()

	// Then
	select {
	case <-termSignal:
	case <-time.After(time.Second):
		t.Fatal("termSignal was not closed after stop()")
	}
}

func newMissingCollaboratorTestCase() []struct {
	name     string
	session  *game.Session
	tracker  *killTracker
	expected string
} {
	return []struct {
		name     string
		session  *game.Session
		tracker  *killTracker
		expected string
	}{
		{"missing session", nil, newKillTracker(0), "game session is required"},
		{"missing tracker", newSessionFixture(game.Config{Speed: 1.0}), nil, "kill tracker is required"},
	}
}
