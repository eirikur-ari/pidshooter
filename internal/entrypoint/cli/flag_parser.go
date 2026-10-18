package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// argumentSplitter partitions command-line arguments into search patterns and
// flag arguments.
type argumentSplitter interface {
	// split returns the search patterns and the flag arguments found in args.
	// It returns an error if the arguments are malformed.
	split(args []string) (patterns, flagArgs []string, err error)
}

// flagParser parses command-line arguments.
type flagParser struct {
	flagSet  *flag.FlagSet
	mapper   *flagMapper
	splitter argumentSplitter
}

// newFlagParser returns a flagParser that recognizes pidshooter's flags.
func newFlagParser() *flagParser {
	flagSet := flag.NewFlagSet("pidshooter", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	return &flagParser{
		flagSet:  flagSet,
		mapper:   newFlagMapper(flagSet),
		splitter: newFlagSplitter(flagSet),
	}
}

// parsePlay returns the play request described by the given arguments. It
// returns an ArgumentError if the arguments are malformed.
func (p *flagParser) parsePlay(args []string) (inbound.PlayRequest, error) {
	patterns, flagArgs, err := p.splitter.split(args)
	if err != nil {
		return inbound.PlayRequest{}, ArgumentError{Cause: err}
	}

	if err := p.flagSet.Parse(flagArgs); err != nil {
		return inbound.PlayRequest{}, ArgumentError{Cause: err}
	}
	// split has already classified every argument as a pattern or as part of
	// flagArgs, so flagSet.Parse can never stop early on a leftover
	// positional. This holds unconditionally today; guarded defensively
	// in case a future change to split ever breaks it.
	if p.flagSet.NArg() > 0 {
		return inbound.PlayRequest{}, ArgumentError{Cause: fmt.Errorf("unexpected argument: %s", p.flagSet.Arg(0))}
	}

	return p.mapper.toPlayRequest(patterns), nil
}
