package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// Program runs a command from command-line arguments.
type Program struct {
	play       inbound.PlayRunner
	list       inbound.ListRunner
	kill       inbound.KillRunner
	errHandler *apperror.Handler
	out        io.Writer
	errOut     io.Writer
}

// NewProgram returns a Program that runs its commands on the given runners, and
// logs errors in the command-line arguments with the given error handler.
func NewProgram(
	play inbound.PlayRunner,
	list inbound.ListRunner,
	kill inbound.KillRunner,
	errHandler *apperror.Handler,
) *Program {
	return &Program{
		play:       play,
		list:       list,
		kill:       kill,
		errHandler: errHandler,
		out:        os.Stdout,
		errOut:     os.Stderr,
	}
}

// Run executes the command the arguments ask for and returns the error it
// ended with, if any. With no arguments or a help flag it prints usage text to
// standard output. When the arguments are malformed it logs the problem, prints
// usage text to standard error and returns an ArgumentError. The error of a
// command is returned as the command reported it, and is not logged again.
func (p *Program) Run(args []string) error {
	root := newRootCommand(p.play, p.list, p.kill).command()
	root.SetArgs(append([]string{}, args...))
	root.SetOut(p.out)
	root.SetErr(p.errOut)

	executed, err := root.ExecuteC()

	var argumentError ArgumentError
	switch {
	case errors.As(err, &argumentError):
		err = p.errHandler.Handle(err)
		p.printUsageText(executed)
	case apperror.CodeInvalidConfig.In(err):
		err = ArgumentError{Cause: err}
		p.printUsageText(executed)
	}

	return err
}

func (p *Program) printUsageText(cmd *cobra.Command) {
	_, _ = fmt.Fprintln(p.errOut)
	cmd.SetOut(p.errOut)
	_ = cmd.Help()
}
