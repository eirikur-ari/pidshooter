package cli

import (
	"flag"
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
  --confirm      Ask for confirmation before killing
  --speed float  Speed multiplier (range: 0.5-5.0) (default 2.0)
  --time int     Time limit in seconds (0 = no limit, max 300) (default 30)
  -h, --help     Show this help
`

// runnerFactory constructs a Runner, deferring any expensive or fallible
// setup until a Runner is actually needed.
type runnerFactory interface {
	Create() (inbound.Runner, error)
}

// CLI is the inbound adapter that translates command-line arguments to application calls.
type CLI struct {
	factory runnerFactory
}

// NewCLI returns a CLI adapter that builds its application service via
// factory. factory is called only once argument parsing and configuration
// validation have both succeeded, so a service that's expensive or
// fallible to construct never affects --help, a parse error, or a
// rejected configuration.
func NewCLI(factory runnerFactory) *CLI {
	return &CLI{factory: factory}
}

// Run parses args and calls the application service.
func (c *CLI) Run(args []string) error {
	err := c.run(args)
	return apperror.LogError(err)
}

func (c *CLI) run(args []string) error {
	if len(args) == 0 {
		_, _ = fmt.Fprint(os.Stdout, usageText)
		return nil
	}

	var confirm bool
	var speed float64
	var timeLimit int
	var help bool
	flagSet := flag.NewFlagSet("pidshooter", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.BoolVar(&confirm, "confirm", false, "")
	flagSet.Float64Var(&speed, "speed", 2.0, "")
	flagSet.IntVar(&timeLimit, "time", 30, "")
	flagSet.BoolVar(&help, "help", false, "")
	flagSet.BoolVar(&help, "h", false, "")

	patterns, flagArgs, err := newFlagSplitter(flagSet).split(args)
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, usageText)
		return err
	}
	if err := flagSet.Parse(flagArgs); err != nil {
		_, _ = fmt.Fprint(os.Stderr, usageText)
		return err
	}
	if help {
		_, _ = fmt.Fprint(os.Stdout, usageText)
		return nil
	}

	cfg, err := config.NewConfig(patterns, confirm, speed, timeLimit)
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, usageText)
		return err
	}

	runner, err := c.factory.Create()
	if err != nil {
		return err
	}
	return runner.Run(cfg)
}
