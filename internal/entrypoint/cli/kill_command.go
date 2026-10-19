package cli

import (
	"github.com/spf13/cobra"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

const killExamples = `  pidshooter kill firefox
  pidshooter kill chrome firefox node`

// killCommand kills the processes its arguments describe.
type killCommand struct {
	runner inbound.KillRunner
	root   *rootFlags
}

// newKillCommand returns a killCommand that kills processes.
func newKillCommand(runner inbound.KillRunner, root *rootFlags) *killCommand {
	return &killCommand{runner: runner, root: root}
}

// command returns the command that kills the processes its arguments
// describe.
func (c *killCommand) command() *cobra.Command {
	return &cobra.Command{
		Use:     "kill <pattern> [pattern2] [pattern3...]",
		Short:   "Kill the processes matching the patterns",
		Example: killExamples,
		RunE: func(_ *cobra.Command, patterns []string) error {
			return c.runner.Run(inbound.KillRequest{
				Patterns: patterns,
				Process:  inbound.ProcessOptions{IncludeRoot: c.root.includeRoot, AllowRoot: c.root.allowRoot},
			})
		},
	}
}
