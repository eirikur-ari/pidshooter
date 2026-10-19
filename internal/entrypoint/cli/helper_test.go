package cli

import "github.com/stretchr/testify/mock"

func runnerMockSetup(playResult error) *MockRunner {
	runner := newRunnerMock()
	runner.On("Run", mock.Anything).Return(playResult)
	return runner
}
