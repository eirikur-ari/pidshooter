package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
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

// Program runs pidshooter from command-line arguments.
type Program struct {
	creator    RunnerCreator
	errHandler *apperror.Handler
	parser     *flagParser
	out        io.Writer
	errOut     io.Writer
}

// RunnerCreator builds a Runner for the given patterns and options, and
// provides the Handler used to log errors.
type RunnerCreator interface {
	// Create returns a Runner that targets processes matching patterns,
	// configured by options. It returns an error if the Runner cannot be built.
	Create(patterns []string, options config.Options) (inbound.Runner, error)
	// ErrHandler returns the Handler used to log errors.
	ErrHandler() *apperror.Handler
}

// NewProgram returns a Program that creates its Runner with creator.
func NewProgram(creator RunnerCreator) *Program {
	return &Program{
		creator:    creator,
		errHandler: creator.ErrHandler(),
		parser:     newFlagParser(),
		out:        os.Stdout,
		errOut:     os.Stderr,
	}
}

// Run runs pidshooter with the given command-line arguments and returns the
// error it ended with, if any. A warning is not treated as an error. With no
// arguments or a help flag it prints usage text to standard output. When the
// arguments are malformed it returns an ArgumentError and prints usage text to
// standard error.
func (p *Program) Run(args []string) error {
	showUsage, err := p.run(args)
	err = p.errHandler.Handle(err)
	p.printUsageText(showUsage, err)
	return err
}

func (p *Program) run(args []string) (showUsage bool, err error) {
	patterns, options, showUsage, err := p.prepare(args)
	if showUsage || err != nil {
		return showUsage, err
	}

	runner, err := p.creator.Create(patterns, options)
	if err != nil {
		return false, err
	}

	if err := runner.Run(); err != nil {
		if apperror.CodeInvalidConfig.In(err) {
			return true, ArgumentError{Cause: err}
		}
		return false, err
	}

	return false, nil
}

// prepare parses the arguments into patterns and config options. showUsage
// reports whether usage text should be shown, whether or not err is set.
func (p *Program) prepare(args []string) (patterns []string, options config.Options, showUsage bool, err error) {
	if len(args) == 0 {
		return nil, config.Options{}, true, nil
	}
	if help(args) {
		return nil, config.Options{}, true, nil
	}

	patterns, options, err = p.parser.parse(args)
	if err != nil {
		return nil, config.Options{}, true, err
	}

	return patterns, options, false, nil
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

// help reports whether the arguments contain -h or --help anywhere.
func help(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}
