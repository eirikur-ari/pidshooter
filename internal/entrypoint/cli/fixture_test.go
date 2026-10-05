package cli

import (
	"flag"
	"io"
)

// newFlagMapperFixture returns a flagMapper backed by a fresh flag.FlagSet,
// along with that FlagSet so tests can parse arguments into it.
func newFlagMapperFixture() (*flagMapper, *flag.FlagSet) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	return newFlagMapper(flagSet), flagSet
}

// newFlagSetFixture returns a FlagSet with one bool flag and one non-bool
// flag registered.
func newFlagSetFixture() *flag.FlagSet {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.Bool("confirm", false, "")
	flagSet.Float64("speed", 2.0, "")
	return flagSet
}
