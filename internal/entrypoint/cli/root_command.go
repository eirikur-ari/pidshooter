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
)

// rootFlags holds the values of the flags that every command accepts. A nil
// field means the flag was not set.
type rootFlags struct {
	includeRoot *bool
	allowRoot   *bool
}

// rootCommand is the top-level command that the other commands run under.
type rootCommand struct {
	play inbound.PlayRunner
	list inbound.ListRunner
	kill inbound.KillRunner

	flags rootFlags
}

// newRootCommand returns a rootCommand that runs its commands on the given
// runners.
func newRootCommand(play inbound.PlayRunner, list inbound.ListRunner, kill inbound.KillRunner) *rootCommand {
	return &rootCommand{play: play, list: list, kill: kill}
}

// command returns the top-level command. Without a command it shows its
// help.
func (c *rootCommand) command() *cobra.Command {
	root := &cobra.Command{
		Use:               "pidshooter",
		Short:             "Process ID Shooter",
		Args:              cobra.ArbitraryArgs,
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: c.mapFlags,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return ArgumentError{Cause: fmt.Errorf("unknown command: %s", args[0])}
		},
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return ArgumentError{Cause: err}
	})
	c.defineFlags(root.PersistentFlags())
	root.AddCommand(
		newPlayCommand(c.play, &c.flags).command(),
		newListCommand(c.list, &c.flags).command(),
		newKillCommand(c.kill, &c.flags).command(),
	)

	return root
}

// defineFlags registers the flags that every command accepts.
func (c *rootCommand) defineFlags(flagSet *pflag.FlagSet) {
	flagSet.Bool(includeRootFlag, false, "Also target root-owned processes")
	flagSet.Bool(iAmRootFlag, false, "Allow running as root")
}

// mapFlags maps the flags that were set on cmd into c.flags. A flag that was
// not set stays nil. It returns an error if a flag cannot be read.
func (c *rootCommand) mapFlags(cmd *cobra.Command, _ []string) error {
	var flags rootFlags
	set := cmd.Flags()

	if set.Changed(includeRootFlag) {
		includeRoot, err := set.GetBool(includeRootFlag)
		if err != nil {
			return err
		}
		flags.includeRoot = new(includeRoot)
	}
	if set.Changed(iAmRootFlag) {
		allowRoot, err := set.GetBool(iAmRootFlag)
		if err != nil {
			return err
		}
		flags.allowRoot = new(allowRoot)
	}

	c.flags = flags

	return nil
}
