package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/adapter/driven/jsonscores"
	"github.com/eirikur-ari/pidshooter/internal/adapter/driven/osprocess"
	"github.com/eirikur-ari/pidshooter/internal/adapter/driven/tcellui"
	"github.com/eirikur-ari/pidshooter/internal/adapter/driving/cli"
	"github.com/eirikur-ari/pidshooter/internal/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	finder := osprocess.NewFinder()
	killer := osprocess.NewKiller()
	store := jsonscores.NewStore()

	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("failed to create screen: %w", err)
	}
	ui := tcellui.New(screen)

	service := app.NewGameService(finder, killer, store, ui, ui)
	return cli.New(service).Run(os.Args[1:])
}
