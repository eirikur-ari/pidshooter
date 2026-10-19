package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestDispatcher_Run_RunsPlayRequestOnPlayRunnerOnly(t *testing.T) {
	// Given
	dispatcher, play, list, kill := newDispatcherFixture()
	request := newPlayRequestFixture()
	play.On("Run", request).Return(nil)

	// When
	err := dispatcher.Run(inbound.RunRequest{Play: &request})

	// Then
	require.NoError(t, err)
	play.AssertNumberOfCalls(t, "Run", 1)
	list.AssertNotCalled(t, "Run", mock.Anything)
	kill.AssertNotCalled(t, "Run", mock.Anything)
}

func TestDispatcher_Run_RunsListRequestOnListRunnerOnly(t *testing.T) {
	// Given
	dispatcher, play, list, kill := newDispatcherFixture()
	request := inbound.ListRequest{Patterns: []string{"target"}}
	list.On("Run", request).Return(nil)

	// When
	err := dispatcher.Run(inbound.RunRequest{List: &request})

	// Then
	require.NoError(t, err)
	list.AssertNumberOfCalls(t, "Run", 1)
	play.AssertNotCalled(t, "Run", mock.Anything)
	kill.AssertNotCalled(t, "Run", mock.Anything)
}

func TestDispatcher_Run_RunsKillRequestOnKillRunnerOnly(t *testing.T) {
	// Given
	dispatcher, play, list, kill := newDispatcherFixture()
	request := inbound.KillRequest{Patterns: []string{"target"}}
	kill.On("Run", request).Return(nil)

	// When
	err := dispatcher.Run(inbound.RunRequest{Kill: &request})

	// Then
	require.NoError(t, err)
	kill.AssertNumberOfCalls(t, "Run", 1)
	play.AssertNotCalled(t, "Run", mock.Anything)
	list.AssertNotCalled(t, "Run", mock.Anything)
}

func TestDispatcher_Run_ReturnsUnknownErrorWhenNoCommandIsSet(t *testing.T) {
	// Given
	var appErr *apperror.Error
	dispatcher, play, list, kill := newDispatcherFixture()

	// When
	err := dispatcher.Run(inbound.RunRequest{})

	// Then
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeUnknown, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.ErrorContains(t, err, "nothing to run")
	play.AssertNotCalled(t, "Run", mock.Anything)
	list.AssertNotCalled(t, "Run", mock.Anything)
	kill.AssertNotCalled(t, "Run", mock.Anything)
}

func TestDispatcher_HandleError_AbsorbsWarningAndReturnsFatal(t *testing.T) {
	// Given
	var appErr *apperror.Error
	dispatcher, _, _, _ := newDispatcherFixture()
	fatal := newFatalError("game session failed")

	// When
	fatalResult := dispatcher.HandleError(fatal)

	// Then
	require.ErrorAs(t, fatalResult, &appErr)
	assert.Same(t, fatal, appErr)
}
