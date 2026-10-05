package cli

import (
	"bytes"
	"flag"
	"io"
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

func newProgramFixture(creator *FakeRunnerCreator) (program *Program, out, errOut *bytes.Buffer) {
	out, errOut = &bytes.Buffer{}, &bytes.Buffer{}
	program = NewProgram(creator)
	program.out, program.errOut = out, errOut
	return program, out, errOut
}
