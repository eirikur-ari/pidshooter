package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application"
	"github.com/eirikur-ari/pidshooter/internal/entrypoint/cli"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/osprocess"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/scorefilestore"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	finder, err := osprocess.NewFinder()
	if err != nil {
		return err
	}
	killer, err := osprocess.NewKiller()
	if err != nil {
		return err
	}
	store := scorefilestore.NewStore()

	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("failed to create screen: %w", err)
	}
	ui := tcellui.New(screen)

	service := application.NewGameRunner(finder, killer, store, ui, ui)
	return cli.New(service).Run(os.Args[1:])
}
