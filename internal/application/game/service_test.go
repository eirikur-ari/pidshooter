package game

import (
	"errors"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestService_Play_RejectsInvalidRequest(t *testing.T) {
	tests := newInvalidPlayRequestTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			service := NewService(nil, nil, nil)
			var appErr *apperror.Error

			// When
			_, err := service.Play(test.request)

			// Then
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, test.expectedCode, appErr.Code)
			assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
			assert.ErrorContains(t, err, test.expectedText)
			assert.Error(t, appErr.Unwrap())
		})
	}
}

func TestService_Play_FailsWhenSessionErrors(t *testing.T) {
	tests := newSessionErrorTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			service := NewService(nil, test.renderer, test.events)
			var appErr *apperror.Error

			// When
			_, err := service.Play(PlayRequest{Speed: 2.0, Processes: []ProcessRequest{newProcessRequestFixture()}})

			// Then
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, apperror.CodeGameFailed, appErr.Code)
			assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
			assert.ErrorContains(t, err, test.expectedText)
			assert.Equal(t, test.expectedCleanupCalls, test.renderer.CleanupCalls)
		})
	}
}

func TestService_runPlaySession_ReturnsErrorWhenSessionIsMissing(t *testing.T) {
	// Given
	service := NewService(nil, nil, nil)

	// When
	_, err := service.runPlaySession(nil, newKillTracker(0))

	// Then
	assert.EqualError(t, err, "game session is required")
}

func TestService_Play_ReturnsResultWhenPlayerQuits(t *testing.T) {
	// Given
	testutil.WarmUpSignalPackage()

	synctest.Test(t, func(t *testing.T) {
		events := testutil.NewFakeInputEventProvider()
		events.Ch <- input.QuitEvent{}
		service := NewService(nil, &testutil.FakeRenderer{}, events)

		// When
		result, err := service.Play(PlayRequest{Speed: 2.0, Processes: []ProcessRequest{newProcessRequestFixture()}})

		// Then
		require.NoError(t, err)
		assert.Equal(t, 0.0, result.Duration)
		assert.Equal(t, 2.0, result.LowestSpeed)
		assert.Equal(t, 0, result.Kills)
		assert.Equal(t, int64(0), result.FreedMem)
		assert.Equal(t, 0, result.Duds)
		assert.Empty(t, result.Errors)
	})
}

func newInvalidPlayRequestTestCase() []struct {
	name         string
	request      PlayRequest
	expectedCode apperror.Code
	expectedText string
} {
	processes := []ProcessRequest{newProcessRequestFixture()}

	return []struct {
		name         string
		request      PlayRequest
		expectedCode apperror.Code
		expectedText string
	}{
		{"speed out of range", PlayRequest{Speed: 99, Processes: processes}, apperror.CodeInvalidConfig, "invalid configuration"},
		{"negative time limit", PlayRequest{Speed: 2.0, TimeLimit: -1, Processes: processes}, apperror.CodeInvalidConfig, "invalid configuration"},
		{"no processes", PlayRequest{Speed: 2.0}, apperror.CodeProcessNotFound, "no processes found"},
	}
}

func newSessionErrorTestCase() []struct {
	name                 string
	renderer             *testutil.FakeRenderer
	events               *testutil.FakeInputEventProvider
	expectedText         string
	expectedCleanupCalls int
} {
	closedEvents := testutil.NewFakeInputEventProvider()
	close(closedEvents.Ch)

	return []struct {
		name                 string
		renderer             *testutil.FakeRenderer
		events               *testutil.FakeInputEventProvider
		expectedText         string
		expectedCleanupCalls int
	}{
		{
			name:         "renderer init fails",
			renderer:     &testutil.FakeRenderer{InitErr: errors.New("display not available")},
			events:       testutil.NewFakeInputEventProvider(),
			expectedText: "renderer initialization failed: display not available",
		},
		{
			name:                 "event channel closes",
			renderer:             &testutil.FakeRenderer{},
			events:               closedEvents,
			expectedText:         "input event channel closed",
			expectedCleanupCalls: 1,
		},
	}
}
