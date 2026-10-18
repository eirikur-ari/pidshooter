package cli

import (
	"bytes"
	"flag"
	"io"

	"github.com/stretchr/testify/mock"
)

func newFlagMapperFixture() (*flagMapper, *flag.FlagSet) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	return newFlagMapper(flagSet), flagSet
}

func newFlagSetFixture() *flag.FlagSet {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.Bool("confirm", false, "")
	flagSet.Float64("speed", 2.0, "")
	return flagSet
}

func newProgramFixture(runner Runner) (program *Program, out, errOut *bytes.Buffer) {
	out, errOut = &bytes.Buffer{}, &bytes.Buffer{}
	program = NewProgram(runner)
	program.out, program.errOut = out, errOut
	return program, out, errOut
}

func newRunnerMock() *MockRunner {
	runner := &MockRunner{}
	runner.On("HandleError", mock.Anything).Return(func(err error) error { return err })
	return runner
}
