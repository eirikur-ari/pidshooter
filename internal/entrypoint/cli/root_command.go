package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

const (
	includeRootFlag = "include-root"
	iAmRootFlag     = "i-am-root"

	defaultMinimumPatterns = 1
)

// rootFlags holds the values of the flags that every command accepts.
type rootFlags struct {
	includeRoot bool
	allowRoot   bool
}

// rootCommand is the top-level command that the other commands run under.
type rootCommand struct {
	play inbound.PlayRunner
	list inbound.ListRunner
	kill inbound.KillRunner

	cmd             *cobra.Command
	flags           rootFlags
	minimumPatterns int
}

// newRootCommand returns a rootCommand that runs its commands on the given
// runners.
func newRootCommand(play inbound.PlayRunner, list inbound.ListRunner, kill inbound.KillRunner) *rootCommand {
	c := &rootCommand{play: play, list: list, kill: kill, minimumPatterns: defaultMinimumPatterns}
	c.cmd = &cobra.Command{
		Use:           "pidshooter",
		Short:         "Process ID Shooter",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          c.showHelp,
	}
	c.cmd.CompletionOptions.DisableDefaultCmd = true
	c.cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return ArgumentError{Cause: err}
	})
	c.defineFlags(c.cmd.PersistentFlags())
	c.addSubcommands()

	return c
}

// showHelp shows the help when there are no arguments, and otherwise returns an
// ArgumentError for the first argument as an unknown command.
func (c *rootCommand) showHelp(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return c.cmd.Help()
	}

	return ArgumentError{Cause: fmt.Errorf("unknown command: %s", args[0])}
}

// addSubcommands adds the commands that run under the top-level command. Each
// of them requires at least one search pattern.
func (c *rootCommand) addSubcommands() {
	for _, sub := range []*cobra.Command{
		newPlayCommand(c.play, c).command(),
		newListCommand(c.list, c).command(),
		newKillCommand(c.kill, c).command(),
	} {
		sub.Args = c.requirePatterns
		c.cmd.AddCommand(sub)
	}
}

// requirePatterns returns an ArgumentError if args holds fewer search patterns
// than the minimum.
func (c *rootCommand) requirePatterns(cmd *cobra.Command, args []string) error {
	if err := cobra.MinimumNArgs(c.minimumPatterns)(cmd, args); err != nil {
		return ArgumentError{Cause: err}
	}

	return nil
}

// defineFlags registers the flags that every command accepts.
func (c *rootCommand) defineFlags(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&c.flags.includeRoot, includeRootFlag, false, "Also target root-owned processes")
	flagSet.BoolVar(&c.flags.allowRoot, iAmRootFlag, false, "Allow running as root")
}

// toProcessOptions returns the process options for the flags that were set in
// flagSet. A flag that was not set stays nil in the options.
func (c *rootCommand) toProcessOptions(flagSet *pflag.FlagSet) inbound.ProcessOptions {
	var options inbound.ProcessOptions

	if flagSet.Changed(includeRootFlag) {
		options.IncludeRoot = new(c.flags.includeRoot)
	}
	if flagSet.Changed(iAmRootFlag) {
		options.AllowRoot = new(c.flags.allowRoot)
	}

	return options
}
