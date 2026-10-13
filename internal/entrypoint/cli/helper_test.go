package cli

import "github.com/stretchr/testify/mock"

func runnerCreatorMockSetup(runResult error) (creator *MockRunnerCreator, runner *MockRunner) {
	runner = &MockRunner{}
	runner.On("Run").Return(runResult)
	creator = &MockRunnerCreator{}
	creator.On("Create", mock.Anything, mock.Anything).Return(runner, nil)
	return creator, runner
}
