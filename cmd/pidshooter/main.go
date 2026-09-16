package main

import (
	"fmt"
	"os"

	"github.com/eirikur-ari/pidshooter/internal/application"
	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/entrypoint/cli"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/filescore"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/osprocess"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/stderr"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/stdout"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	logger := stderr.NewLogger()

	// These two adapters are constructed before application.NewRunner, so
	// there's no apperror.Handler yet to route their failures through —
	// they're the one place in the program that logs directly.
	proc, err := osprocess.NewProcess()
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	store, err := filescore.NewFileScore()
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		err = fmt.Errorf("failed to create screen: %w", err)
		logger.Error(err.Error())
		return err
	}
	ui := tcellui.NewTUI(screen)
	reporter := stdout.NewScoreReporter()

	runner := application.NewRunner(proc, store, reporter, ui, ui, logger)
	return cli.NewCLI(runner, logger).Run(os.Args[1:])
}
