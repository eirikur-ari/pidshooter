// Package cli implements the delivery adapter that parses os.Args and calls the application service.
package cli

import (
	"errors"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// CLI is the entrypoint adapter that translates command-line arguments to application calls.
type CLI struct {
	service   inbound.Runner
	logger    logger
	out       io.Writer
	errOut    io.Writer
	confirm   bool
	speed     float64
	timeLimit int
}

// logger is the minimal reporting capability this adapter needs.
type logger interface {
	Error(msg string)
}

// NewCLI returns a CLI adapter wrapping the given application service.
func NewCLI(service inbound.Runner, logger logger) *CLI {
	return &CLI{service: service, logger: logger, out: os.Stdout, errOut: os.Stderr}
}

// Run parses args and calls the application service.
func (c *CLI) Run(args []string) error {
	cmd := c.buildCommand()
	cmd.SetArgs(args)
	return c.report(cmd.Execute())
}

// report logs err if it's native to this adapter — such as a malformed
// flag — rather than already reported by the application layer, then
// returns it unchanged. This relies on an invariant the application layer
// must uphold: every *apperror.Error it constructs is routed through
// apperror.Handler.Handle (and therefore already logged) before it can
// reach this adapter. A path that returns a raw *apperror.Error without
// going through Handle first would be silently un-logged here.
func (c *CLI) report(err error) error {
	var appErr *apperror.Error
	if err != nil && !errors.As(err, &appErr) {
		c.logger.Error(err.Error())
	}
	return err
}

func (c *CLI) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pidshooter <pattern> [pattern2] [pattern3...]",
		Short: "Process ID Shooter",
		Long:  `Hunt running processes by name and kill them in a terminal shooter game.`,
		Example: `  pidshooter firefox
  pidshooter chrome firefox node
  pidshooter firefox --confirm
  pidshooter node --speed=2.5 --time=60
  pidshooter node --time=0`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          c.play,
	}

	cmd.SetOut(c.out)
	cmd.SetErr(c.errOut)

	cmd.Flags().BoolVar(&c.confirm, "confirm", false, "Ask for confirmation before killing")
	cmd.Flags().Float64Var(&c.speed, "speed", 2.0, "Speed multiplier (range: 0.1-5.0)")
	cmd.Flags().IntVar(&c.timeLimit, "time", 30, "Time limit in seconds (0 = no limit)")

	return cmd
}

func (c *CLI) play(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && cmd.Flags().NFlag() == 0 {
		return cmd.Help()
	}

	return c.service.Run(inbound.Config{
		Patterns:    args,
		ConfirmMode: c.confirm,
		Speed:       c.speed,
		TimeLimit:   c.timeLimit,
	})
}
