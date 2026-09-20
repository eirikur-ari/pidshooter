package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/util"
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
  --speed float  Speed multiplier (range: 0.1-5.0) (default 2.0)
  --time int     Time limit in seconds (0 = no limit) (default 30)
  -h, --help     Show this help
`

// CLI is the inbound adapter that translates command-line arguments to application calls.
type CLI struct {
	service inbound.Runner
}

// NewCLI returns a CLI adapter wrapping the given application service.
func NewCLI(service inbound.Runner) *CLI {
	return &CLI{service: service}
}

// Run parses args and calls the application service.
func (c *CLI) Run(args []string) error {
	return c.report(c.play(args))
}

func (c *CLI) play(args []string) error {
	if len(args) == 0 {
		_, _ = fmt.Fprint(os.Stdout, usageText)
		return nil
	}

	var confirm bool
	var speed float64
	var timeLimit int
	var help bool
	fs := flag.NewFlagSet("pidshooter", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&confirm, "confirm", false, "Ask for confirmation before killing")
	fs.Float64Var(&speed, "speed", 2.0, "Speed multiplier (range: 0.1-5.0)")
	fs.IntVar(&timeLimit, "time", 30, "Time limit in seconds (0 = no limit)")
	fs.BoolVar(&help, "help", false, "Show this help")
	fs.BoolVar(&help, "h", false, "Show this help")

	patterns, flagArgs, err := splitArgs(fs, args)
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, usageText)
		return err
	}
	if err := fs.Parse(flagArgs); err != nil {
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

	return c.service.Run(cfg)
}

// report logs err through the adapter's logger, unless err is (or wraps) an
// *apperror.Error, which is assumed to already be logged.
func (c *CLI) report(err error) error {
	var appErr *apperror.Error
	if err != nil && !errors.As(err, &appErr) {
		util.NewLogger().Error(err.Error())
	}
	return err
}

// splitArgs partitions args into search patterns and flag tokens. Any
// argument that isn't a flag registered on fs (or that flag's value) is
// treated as a pattern, regardless of where it falls among the flags.
func splitArgs(fs *flag.FlagSet, args []string) (patterns, flagArgs []string, err error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "-" || !strings.HasPrefix(a, "-") {
			patterns = append(patterns, a)
			continue
		}

		name, hasValue := flagNameAndValue(a)
		f := fs.Lookup(name)
		if f == nil {
			return nil, nil, fmt.Errorf("flag provided but not defined: -%s", name)
		}
		flagArgs = append(flagArgs, a)
		if hasValue {
			continue
		}

		if bf, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
			continue
		}
		if i+1 >= len(args) {
			return nil, nil, fmt.Errorf("flag needs an argument: -%s", name)
		}
		i++
		flagArgs = append(flagArgs, args[i])
	}
	return patterns, flagArgs, nil
}

// flagNameAndValue extracts a flag's name from a "-name" or "--name=value"
// token, reporting whether the value was inlined with "=".
func flagNameAndValue(a string) (name string, hasValue bool) {
	name = strings.TrimPrefix(strings.TrimPrefix(a, "-"), "-")
	if before, _, ok := strings.Cut(name, "="); ok {
		return before, true
	}
	return name, false
}
