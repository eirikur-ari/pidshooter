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

// newFlagParser returns a flagParser that recognizes the supported flags.
func newFlagParser() *flagParser {
	flagSet := flag.NewFlagSet("pidshooter", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	return &flagParser{
		flagSet:  flagSet,
		mapper:   newFlagMapper(flagSet),
		splitter: newFlagSplitter(flagSet),
	}
}

// parseRequest returns the run request described by the arguments of cmd. It
// returns an ArgumentError if cmd is not a command or its arguments are
// malformed.
func (p *flagParser) parseRequest(cmd command) (inbound.RunRequest, error) {
	switch {
	case cmd.Play():
		request, err := p.parsePlay(cmd.Arguments())
		return inbound.RunRequest{Play: &request}, err
	case cmd.List():
		request, err := p.parseList(cmd.Arguments())
		return inbound.RunRequest{List: &request}, err
	case cmd.Kill():
		request, err := p.parseKill(cmd.Arguments())
		return inbound.RunRequest{Kill: &request}, err
	default:
		return inbound.RunRequest{}, ArgumentError{Cause: fmt.Errorf("unknown command: %s", cmd.Name())}
	}
}

// parsePlay returns the play request described by the given arguments. It
// returns an ArgumentError if the arguments are malformed.
func (p *flagParser) parsePlay(args []string) (inbound.PlayRequest, error) {
	patterns, err := p.parse(args)
	if err != nil {
		return inbound.PlayRequest{}, err
	}

	return p.mapper.toPlayRequest(patterns), nil
}

// parseList returns the list request described by the given arguments. It
// returns an ArgumentError if the arguments are malformed.
func (p *flagParser) parseList(args []string) (inbound.ListRequest, error) {
	patterns, err := p.parse(args)
	if err != nil {
		return inbound.ListRequest{}, err
	}

	return inbound.ListRequest{Patterns: patterns}, nil
}

// parseKill returns the kill request described by the given arguments. It
// returns an ArgumentError if the arguments are malformed.
func (p *flagParser) parseKill(args []string) (inbound.KillRequest, error) {
	patterns, err := p.parse(args)
	if err != nil {
		return inbound.KillRequest{}, err
	}

	return inbound.KillRequest{Patterns: patterns}, nil
}

// parse returns the search patterns in args and applies its flags to the flag
// set. It returns an ArgumentError if the arguments are malformed.
func (p *flagParser) parse(args []string) ([]string, error) {
	patterns, flagArgs, err := p.splitter.split(args)
	if err != nil {
		return nil, ArgumentError{Cause: err}
	}

	if err := p.flagSet.Parse(flagArgs); err != nil {
		return nil, ArgumentError{Cause: err}
	}
	// split has already classified every argument as a pattern or as part of
	// flagArgs, so flagSet.Parse can never stop early on a leftover
	// positional. This holds unconditionally today; guarded defensively
	// in case a future change to split ever breaks it.
	if p.flagSet.NArg() > 0 {
		return nil, ArgumentError{Cause: fmt.Errorf("unexpected argument: %s", p.flagSet.Arg(0))}
	}

	return patterns, nil
}
