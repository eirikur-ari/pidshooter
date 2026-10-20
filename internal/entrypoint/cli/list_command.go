package cli

import (
	"github.com/spf13/cobra"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

const listExamples = `  pidshooter list firefox
  pidshooter list chrome firefox node`

// listCommand lists the processes its arguments describe.
type listCommand struct {
	runner inbound.ListRunner
	root   *rootCommand
}

// newListCommand returns a listCommand that lists processes.
func newListCommand(runner inbound.ListRunner, root *rootCommand) *listCommand {
	return &listCommand{runner: runner, root: root}
}

// command returns the command that lists the processes its arguments
// describe.
func (c *listCommand) command() *cobra.Command {
	return &cobra.Command{
		Use:     "list <pattern> [pattern2] [pattern3...]",
		Short:   "List the processes matching the patterns",
		Example: listExamples,
		RunE: func(cmd *cobra.Command, patterns []string) error {
			return c.runner.Run(inbound.ListRequest{
				Patterns: patterns,
				Process:  c.root.toProcessOptions(cmd.Flags()),
			})
		},
	}
}
