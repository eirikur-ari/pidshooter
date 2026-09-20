package main

import (
	"os"

	"github.com/eirikur-ari/pidshooter/internal/composition"
	"github.com/eirikur-ari/pidshooter/internal/entrypoint/cli"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	return cli.NewCLI(composition.NewRunnerFactory()).Run(os.Args[1:])
}
