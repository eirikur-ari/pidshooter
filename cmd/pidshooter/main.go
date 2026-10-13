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

// run runs pidshooter with the command-line arguments and returns the error it
// ended with, if any.
func run() error {
	creator := composition.NewRunnerCreator()
	program := cli.NewProgram(creator)
	return program.Run(os.Args[1:])
}

// exitCode returns the process exit status for err: 2 when the command-line
// arguments were malformed, 1 for any other error, and 0 when there is none.
func exitCode(err error) int {
	var argumentError cli.ArgumentError

	if errors.As(err, &argumentError) {
		return 2
	}

	if err != nil {
		return 1
	}

	return 0
}
