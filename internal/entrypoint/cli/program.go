package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

const usageText = `Process ID Shooter

Find running processes by name and kill them.

Usage:
  pidshooter <command> <pattern> [pattern2] [pattern3...] [flags]

Commands:
  play             Play the game: shoot the processes matching the patterns

Examples:
  pidshooter play firefox
  pidshooter play chrome firefox node
  pidshooter play firefox --confirm
  pidshooter play node --speed=2.5 --time=60
  pidshooter play node --time=0

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

// Runner runs the command a request describes.
type Runner interface {
	// Run runs the command described by request, returning an error if it
	// could not complete.
	Run(request inbound.RunRequest) error
	// HandleError logs err at the level its severity calls for. It returns nil if
	// the error is absorbed, and err unchanged otherwise.
	HandleError(err error) error
}

// Program runs a command from command-line arguments.
type Program struct {
	runner Runner
	parser *flagParser
	out    io.Writer
	errOut io.Writer
}

// NewProgram returns a Program that executes commands with runner and logs
// errors with the error handler of runner.
func NewProgram(runner Runner) *Program {
	return &Program{
		runner: runner,
		parser: newFlagParser(),
		out:    os.Stdout,
		errOut: os.Stderr,
	}
}

// Run executes the command the arguments ask for and returns the error it
// ended with, if any. A warning is not treated as an error. With no arguments
// or a help flag it prints usage text to standard output. When the first
// argument is not a command, or the arguments are malformed, it returns an
// ArgumentError and prints usage text to standard error.
func (p *Program) Run(args []string) error {
	showUsage, err := p.run(newCommand(args))
	err = p.runner.HandleError(err)
	p.printUsageText(showUsage, err)
	return err
}

func (p *Program) run(cmd command) (showUsage bool, err error) {
	if cmd.Help() {
		return true, nil
	}

	request, err := p.parser.parseRequest(cmd)
	if err != nil {
		return true, err
	}

	if err := p.runner.Run(request); err != nil {
		if apperror.CodeInvalidConfig.In(err) {
			return true, ArgumentError{Cause: err}
		}
		return false, err
	}

	return false, nil
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
