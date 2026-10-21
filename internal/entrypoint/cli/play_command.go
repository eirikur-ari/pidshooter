package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

const (
	confirmFlag = "confirm"
	speedFlag   = "speed"
	timeFlag    = "time"
)

const playLongDescription = `Process ID Shooter

Find running processes by name and kill them.

Config file:
  Defaults for the flags are read from ~/.config/pidshooter/config.yaml
  (or $XDG_CONFIG_HOME/pidshooter/config.yaml, if set), when present. A flag
  passed on the command line always overrides the config file for that run —
  e.g. --include-root=false overrides a persisted "include_root: true".`

const playExamples = `  pidshooter play firefox
  pidshooter play chrome firefox node
  pidshooter play firefox --confirm
  pidshooter play node --speed=2.5 --time=60
  pidshooter play node --time=0`

// playFlags holds the values of the flags of the play command.
type playFlags struct {
	confirm   bool
	speed     float64
	timeLimit int
}

// playCommand plays the game session its arguments describe.
type playCommand struct {
	runner inbound.PlayRunner
	flags  playFlags
	root   *rootCommand
}

// newPlayCommand returns a playCommand that plays game sessions.
func newPlayCommand(runner inbound.PlayRunner, root *rootCommand) *playCommand {
	return &playCommand{runner: runner, root: root}
}

// command returns the command that plays the game session its arguments
// describe.
func (c *playCommand) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "play <pattern> [pattern2] [pattern3...]",
		Short:   "Find running processes by name and kill them",
		Long:    playLongDescription,
		Example: playExamples,
		RunE: func(cmd *cobra.Command, patterns []string) error {
			return c.runner.Run(c.toRequest(cmd.Flags(), patterns))
		},
	}
	c.defineFlags(cmd.Flags())

	return cmd
}

// defineFlags registers the flags of the command.
func (c *playCommand) defineFlags(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&c.flags.confirm, confirmFlag, false, "Ask for confirmation before killing")
	flagSet.Float64Var(&c.flags.speed, speedFlag, 0, "Speed multiplier (range: 0.5-5.0); defaults to the config file value, or 2.0 if unset")
	flagSet.IntVar(&c.flags.timeLimit, timeFlag, 0, "Time limit in seconds (0 = no limit, max 300); defaults to the config file value, or 30 if unset")
}

// toRequest returns the play request for the patterns and the flags that were
// set. A flag that was not set stays nil in the request.
func (c *playCommand) toRequest(flagSet *pflag.FlagSet, patterns []string) inbound.PlayRequest {
	request := inbound.PlayRequest{
		Patterns: patterns,
		Process:  c.root.toProcessOptions(flagSet),
	}

	if flagSet.Changed(confirmFlag) {
		request.Game.ConfirmMode = new(c.flags.confirm)
	}
	if flagSet.Changed(speedFlag) {
		request.Game.Speed = new(c.flags.speed)
	}
	if flagSet.Changed(timeFlag) {
		request.Game.TimeLimit = new(c.flags.timeLimit)
	}

	return request
}
