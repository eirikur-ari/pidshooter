package cli

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func returnFromEveryRunner(fixture programFixture, err error) {
	fixture.play.On("Run", mock.Anything).Return(err)
	fixture.list.On("Run", mock.Anything).Return(err)
	fixture.kill.On("Run", mock.Anything).Return(err)
}

func assertNoRunnerRan(t mock.TestingT, play *MockPlayRunner, list *MockListRunner, kill *MockKillRunner) {
	play.AssertNotCalled(t, "Run", mock.Anything)
	list.AssertNotCalled(t, "Run", mock.Anything)
	kill.AssertNotCalled(t, "Run", mock.Anything)
}

func executeCommand(cmd *cobra.Command, args []string) error {
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return cmd.Execute()
}
