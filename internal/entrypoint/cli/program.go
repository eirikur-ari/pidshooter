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
	buildVersion string
	play         inbound.PlayRunner
	list         inbound.ListRunner
	kill         inbound.KillRunner
	errHandler   *apperror.Handler
	out          io.Writer
	errOut       io.Writer
}

// NewProgram returns a Program that reports the given build version, runs its
// commands on the given runners, and logs errors in the command-line arguments
// with the given error handler.
func NewProgram(
	buildVersion string,
	play inbound.PlayRunner,
	list inbound.ListRunner,
	kill inbound.KillRunner,
	errHandler *apperror.Handler,
) *Program {
	return &Program{
		buildVersion: buildVersion,
		play:         play,
		list:         list,
		kill:         kill,
		errHandler:   errHandler,
		out:          os.Stdout,
		errOut:       os.Stderr,
	}
}

// Run executes the command the arguments ask for and returns the error it
// ended with, if any. With no arguments or a help flag it prints usage text to
// standard output. When the arguments are malformed it logs the problem, prints
// usage text to standard error and returns an ArgumentError. A command error
// classified as invalid configuration is returned as an ArgumentError with
// usage text printed to standard error. Any other error of a command is
// returned as the command reported it, and is not logged again.
func (p *Program) Run(args []string) error {
	root := newRootCommand(p.buildVersion, p.play, p.list, p.kill).cmd
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

// printUsageText prints a blank line and the usage text of cmd to standard
// error.
func (p *Program) printUsageText(cmd *cobra.Command) {
	_, _ = fmt.Fprintln(p.errOut)
	cmd.SetOut(p.errOut)
	_ = cmd.Help()
}
