package main

import (
	"errors"
	"os"

	"github.com/eirikur-ari/pidshooter/internal/composition"
	"github.com/eirikur-ari/pidshooter/internal/entrypoint/cli"
)

func main() {
	err := run()
	code := exitCode(err)
	os.Exit(code)
}

func run() error {
	factory := composition.NewRunnerFactory()
	program := cli.NewProgram(factory)
	return program.Run(os.Args[1:])
}

func exitCode(err error) int {
	var argErr cli.ArgumentError

	if errors.As(err, &argErr) {
		return 2
	}

	if err != nil {
		return 1
	}

	return 0
}
