package cli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

const usageText = `Process ID Shooter

Find running processes by name and kill them.

Usage:
  pidshooter <pattern> [pattern2] [pattern3...] [flags]

Examples:
  pidshooter firefox
  pidshooter chrome firefox node
  pidshooter firefox --confirm
  pidshooter node --speed=2.5 --time=60
  pidshooter node --time=0

Flags:
  --confirm        Ask for confirmation before killing
  --speed float    Speed multiplier (range: 0.5-5.0); defaults to the config file value, or 2.0 if unset
  --time int       Time limit in seconds (0 = no limit, max 300); defaults to the config file value, or 30 if unset
  --include-root   Also target root-owned processes (always allowed when running as root)
  --i-am-root      Allow running as root (refused by default: as root, every process on the machine becomes a target)
  -h, --help       Show this help

Config file:
  Defaults for the flags above are read from ~/.config/pidshooter/config.yaml
  (or $XDG_CONFIG_HOME/pidshooter/config.yaml, if set), when present. A flag
  passed on the command line always overrides the config file for that run —
  e.g. --include-root=false overrides a persisted "include_root: true".
`

// Program translates command-line arguments to application calls.
type Program struct {
	creator    runnerCreator
	errHandler *apperror.Handler
	out        io.Writer
	errOut     io.Writer
}

// runnerCreator constructs a Runner, deferring any expensive or fallible
// setup until a Runner is actually needed, and provides a Handler for
// logging errors.
type runnerCreator interface {
	Create() (inbound.Runner, error)
	ErrHandler() *apperror.Handler
}

// NewProgram returns a Program that builds its application service via creator.
func NewProgram(creator runnerCreator) *Program {
	return &Program{creator: creator, errHandler: creator.ErrHandler(), out: os.Stdout, errOut: os.Stderr}
}

// Run parses args and calls the application service.
func (p *Program) Run(args []string) error {
	showUsage, err := p.run(args)
	err = p.errHandler.Handle(err)
	p.printUsageText(showUsage, err)
	return err
}

func (p *Program) run(args []string) (showUsage bool, err error) {
	req, showUsage, err := p.prepare(args)
	if showUsage || err != nil {
		return showUsage, err
	}

	runner, err := p.creator.Create()
	if err != nil {
		return false, err
	}

	if err := runner.Run(req); err != nil {
		if apperror.CodeInvalidConfig.In(err) {
			return true, ArgumentError{Cause: err}
		}
		return false, err
	}

	return false, nil
}

// prepare parses args into a RunRequest. showUsage reports whether usage
// text should be shown, independent of whether err is also set.
func (p *Program) prepare(args []string) (req inbound.RunRequest, showUsage bool, err error) {
	if len(args) == 0 {
		return inbound.RunRequest{}, true, nil
	}
	if help(args) {
		return inbound.RunRequest{}, true, nil
	}

	flagSet := flag.NewFlagSet("pidshooter", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	mapper := newFlagMapper(flagSet)

	patterns, flagArgs, err := newFlagSplitter(flagSet).split(args)
	if err != nil {
		return inbound.RunRequest{}, true, ArgumentError{Cause: err}
	}

	if err := flagSet.Parse(flagArgs); err != nil {
		return inbound.RunRequest{}, true, ArgumentError{Cause: err}
	}
	// split has already classified every token as a pattern or as part of
	// flagArgs, so flagSet.Parse can never stop early on a leftover
	// positional. This holds unconditionally today; guarded defensively
	// in case a future change to split ever breaks it.
	if flagSet.NArg() > 0 {
		return inbound.RunRequest{}, true, ArgumentError{Cause: fmt.Errorf("unexpected argument: %s", flagSet.Arg(0))}
	}

	return mapper.toRunRequest(patterns), false, nil
}

func (p *Program) printUsageText(printUsage bool, err error) {
	if !printUsage {
		return
	}
	out := p.out
	if err != nil {
		out = p.errOut
		_, _ = fmt.Fprintln(out)
	}
	_, _ = fmt.Fprint(out, usageText)
}

// help reports whether args contains -h or --help, regardless of its
// position among other arguments or any parse error elsewhere in args.
func help(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}
