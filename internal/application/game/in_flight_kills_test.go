package game

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

func TestInFlightKills_start_ReportsKillerResultToTracker(t *testing.T) {
	tests := newKillerResultTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			killer := &MockProcessKiller{}
			killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(test.err)
			tracker := newKillTracker(0)
			kills := newInFlightKills(killer, defaultKillGracePeriod, tracker)

			// When
			kills.start(newTargetFixture())
			kills.awaitRemaining()

			// Then
			assert.Equal(t, test.expected.kills, tracker.score.kills)
			assert.Len(t, tracker.duds, test.expected.duds)
			assert.Len(t, tracker.failure.failures, test.expected.failures)
		})
	}
}

func TestInFlightKills_start_PassesTargetDetailsToKiller(t *testing.T) {
	tests := newKillArgumentsTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			killer := &MockProcessKiller{}
			killer.On("Kill", test.expected.pid, test.expected.name, test.expected.protected).Return(nil)
			kills := newInFlightKills(killer, defaultKillGracePeriod, newKillTracker(0))

			// When
			kills.start(test.target)
			kills.awaitRemaining()

			// Then
			killer.AssertExpectations(t)
		})
	}
}

func TestInFlightKills_applyFinished_AppliesFinishedKills(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		killer := &MockProcessKiller{}
		killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		tracker := newKillTracker(0)
		kills := newInFlightKills(killer, defaultKillGracePeriod, tracker)
		kills.start(newTargetFixture())
		synctest.Wait()

		// When
		kills.applyFinished()

		// Then
		assert.Equal(t, 1, tracker.score.kills)
	})
}

func TestInFlightKills_applyFinished_ReturnsWithoutBlockingWhenNothingHasFinished(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		killer := &MockProcessKiller{}
		killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(nil).After(time.Hour)
		tracker := newKillTracker(0)
		kills := newInFlightKills(killer, defaultKillGracePeriod, tracker)
		kills.start(newTargetFixture())
		defer kills.waitGroup.Wait()
		synctest.Wait()

		// When
		kills.applyFinished()

		// Then
		assert.Equal(t, 0, tracker.score.kills)
	})
}

func TestInFlightKills_awaitRemaining_WaitsForSlowKill(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		const killDelay = 300 * time.Millisecond
		killer := &MockProcessKiller{}
		killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(nil).After(killDelay)
		tracker := newKillTracker(0)
		kills := newInFlightKills(killer, time.Second, tracker)
		kills.start(newTargetFixture())
		start := time.Now()

		// When
		kills.awaitRemaining()

		// Then
		assert.Equal(t, killDelay, time.Since(start))
		assert.Equal(t, 1, tracker.score.kills)
	})
}

func TestInFlightKills_awaitRemaining_GivesUpOnceGracePeriodElapses(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		const gracePeriod = 20 * time.Millisecond
		killer := &MockProcessKiller{}
		killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(nil).After(time.Hour)
		tracker := newKillTracker(0)
		kills := newInFlightKills(killer, gracePeriod, tracker)
		kills.start(newTargetFixture())
		defer kills.waitGroup.Wait()
		start := time.Now()

		// When
		kills.awaitRemaining()

		// Then
		assert.Equal(t, gracePeriod, time.Since(start))
		assert.Equal(t, 0, tracker.score.kills)
	})
}

func TestInFlightKills_close_ReleasesKillsWaitingToReportTheirResult(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Given
		killer := &MockProcessKiller{}
		killer.On("Kill", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		kills := newInFlightKills(killer, defaultKillGracePeriod, newKillTracker(0))
		target := newTargetFixture()
		kickedOff := cap(kills.results) + 1
		for range kickedOff {
			kills.start(target)
		}
		synctest.Wait()

		// When
		kills.close()
		kills.waitGroup.Wait()

		// Then
		killer.AssertNumberOfCalls(t, "Kill", kickedOff)
	})
}

func newKillerResultTestCase() []struct {
	name     string
	err      error
	expected struct{ kills, duds, failures int }
} {
	type counts = struct{ kills, duds, failures int }

	return []struct {
		name     string
		err      error
		expected counts
	}{
		{"kill succeeded", nil, counts{kills: 1}},
		{"kill failed", apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, "refusing to kill PID 100", nil), counts{failures: 1}},
		{"process already gone drops the not found error", apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, "PID 100 already exited", nil), counts{duds: 1}},
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
