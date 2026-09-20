package main

import (
	"fmt"
	"os"

	"github.com/eirikur-ari/pidshooter/internal/application"
	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/entrypoint/cli"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/console"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/filescore"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/osprocess"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	logger := util.NewLogger()

	// A failure here means the program can't run at all, so it's logged
	// and the program exits immediately rather than continuing on.
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
	reporter := console.NewScoreReporter()

	runner := application.NewRunner(proc, store, reporter, ui, ui)
	return cli.NewCLI(runner).Run(os.Args[1:])
}
