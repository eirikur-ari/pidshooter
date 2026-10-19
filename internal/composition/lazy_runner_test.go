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
	builds := 0
	lazy := &lazyRunner[inbound.PlayRequest, inbound.PlayRunner]{
		build: func() (inbound.PlayRunner, error) {
			builds++
			return runner, nil
		},
	}
	buildsBeforeRun := builds

	// When
	err := lazy.Run(request)

	// Then
	require.NoError(t, err)
	assert.Equal(t, 0, buildsBeforeRun)
	assert.Equal(t, 1, builds)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestLazyRunner_Run_ReturnsBuildErrorWithoutRunningRequest(t *testing.T) {
	// Given
	failure := errors.New("cannot set up")
	lazy := &lazyRunner[inbound.PlayRequest, inbound.PlayRunner]{
		build: func() (inbound.PlayRunner, error) { return nil, failure },
	}

	// When
	err := lazy.Run(inbound.PlayRequest{})

	// Then
	assert.Same(t, failure, err)
}

func TestLazyRunner_Run_ReturnsErrorOfBuiltRunner(t *testing.T) {
	// Given
	failure := errors.New("session failed")
	runner := &MockPlayRunner{}
	runner.On("Run", inbound.PlayRequest{}).Return(failure)
	lazy := &lazyRunner[inbound.PlayRequest, inbound.PlayRunner]{
		build: func() (inbound.PlayRunner, error) { return runner, nil },
	}

	// When
	err := lazy.Run(inbound.PlayRequest{})

	// Then
	assert.Same(t, failure, err)
}
