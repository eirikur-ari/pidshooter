package main

import (
	"fmt"
	"os"

	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/gdamore/tcell/v2"

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
	proc, err := osprocess.NewProcess()
	if err != nil {
		return err
	}
	store := scorefilestore.NewStore()

	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("failed to create screen: %w", err)
	}
	ui := tcellui.NewUI(screen)

	svc := game.NewService(proc, store, ui, ui)
	return cli.NewCLI(svc).Run(os.Args[1:])
}
