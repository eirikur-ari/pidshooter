package game

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/input"
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

func TestService_runPlaySession_ReturnsErrorWhenSessionOrTrackerIsMissing(t *testing.T) {
	tests := newMissingCollaboratorTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			service := NewService(nil, nil, nil)

			// When
			_, err := service.runPlaySession(test.session, test.tracker)

			// Then
			assert.EqualError(t, err, test.expected)
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
