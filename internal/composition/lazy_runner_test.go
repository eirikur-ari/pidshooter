package composition

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestLazyRunner_Run_BuildsRunnerOnlyWhenRun(t *testing.T) {
	// Given
	request := inbound.PlayRequest{Patterns: []string{"proc"}}
	runner := &MockPlayRunner{}
	runner.On("Run", request).Return(nil)
	handler, _ := newHandlerFixture()
	build, builds := newCountingBuildFixture(runner)
	lazy := newLazyRunnerFixture(build, handler)
	buildsBeforeRun := *builds

	// When
	err := lazy.Run(request)

	// Then
	require.NoError(t, err)
	assert.Equal(t, 0, buildsBeforeRun)
	assert.Equal(t, 1, *builds)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestLazyRunner_Run_HandlesBuildErrorAndReturnsItWithoutRunningRequest(t *testing.T) {
	// Given
	failure := errors.New("cannot set up")
	handler, logger := newHandlerFixture()
	lazy := newLazyRunnerFixture(newBuildFixture(nil, failure), handler)

	// When
	err := lazy.Run(inbound.PlayRequest{})

	// Then
	assert.Same(t, failure, err)
	assert.Equal(t, []string{"cannot set up"}, logger.Errored)
}

func TestLazyRunner_Run_ReturnsBuildErrorEvenWhenItsSeverityIsAbsorbed(t *testing.T) {
	// Given
	warning := newWarningFixture("cannot set up")
	handler, logger := newHandlerFixture()
	lazy := newLazyRunnerFixture(newBuildFixture(nil, warning), handler)

	// When
	err := lazy.Run(inbound.PlayRequest{})

	// Then
	assert.Same(t, warning, err)
	assert.Equal(t, []string{"cannot set up"}, logger.Warned)
}

func TestLazyRunner_Run_ReturnsErrorOfBuiltRunnerWithoutHandlingIt(t *testing.T) {
	// Given
	failure := errors.New("session failed")
	runner := &MockPlayRunner{}
	runner.On("Run", inbound.PlayRequest{}).Return(failure)
	handler, logger := newHandlerFixture()
	lazy := newLazyRunnerFixture(newBuildFixture(runner, nil), handler)

	// When
	err := lazy.Run(inbound.PlayRequest{})

	// Then
	assert.Same(t, failure, err)
	assert.Empty(t, logger.Errored)
	assert.Empty(t, logger.Warned)
}
