package cli

import (
	"flag"
	"io"
)

// newFlagMapperFixture returns a flagMapper backed by a fresh flag.FlagSet, for
// tests to Parse args into before calling toConfigOptions.
func newFlagMapperFixture() (*flagMapper, *flag.FlagSet) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	return newFlagMapper(flagSet), flagSet
}
