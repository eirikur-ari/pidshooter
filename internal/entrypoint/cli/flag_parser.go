package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
)

// flagParser parses command-line arguments into search patterns and config
// options.
type flagParser struct {
	flagSet  *flag.FlagSet
	mapper   *flagMapper
	splitter argumentSplitter
}

// argumentSplitter partitions command-line arguments into search patterns and
// flag arguments.
type argumentSplitter interface {
	// split returns the search patterns and the flag arguments found in args.
	// It returns an error if the arguments are malformed.
	split(args []string) (patterns, flagArgs []string, err error)
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

// parse returns the search patterns and config options found in args. It
// returns an ArgumentError if the arguments are malformed.
func (p *flagParser) parse(args []string) (patterns []string, options config.Options, err error) {
	patterns, flagArgs, err := p.splitter.split(args)
	if err != nil {
		return nil, config.Options{}, ArgumentError{Cause: err}
	}

	if err := p.flagSet.Parse(flagArgs); err != nil {
		return nil, config.Options{}, ArgumentError{Cause: err}
	}
	// split has already classified every argument as a pattern or as part of
	// flagArgs, so flagSet.Parse can never stop early on a leftover
	// positional. This holds unconditionally today; guarded defensively
	// in case a future change to split ever breaks it.
	if p.flagSet.NArg() > 0 {
		return nil, config.Options{}, ArgumentError{Cause: fmt.Errorf("unexpected argument: %s", p.flagSet.Arg(0))}
	}

	return patterns, p.mapper.toConfigOptions(), nil
}
