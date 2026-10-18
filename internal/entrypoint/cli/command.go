package cli

// command is what the command-line arguments ask pidshooter to do.
type command struct {
	args []string
}

// newCommand returns the command that args ask for.
func newCommand(args []string) command {
	return command{args: args}
}

// Help reports whether the command is to show the usage text: no arguments, or
// -h or --help anywhere.
func (c command) Help() bool {
	if len(c.args) == 0 {
		return true
	}

	for _, arg := range c.args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

// Play reports whether the command is to play the game.
func (c command) Play() bool {
	return !c.Help() && c.args[0] == "play"
}

// Name returns the first argument, or "" if there are no arguments.
func (c command) Name() string {
	if len(c.args) == 0 {
		return ""
	}
	return c.args[0]
}

// Arguments returns the arguments that follow the command name.
func (c command) Arguments() []string {
	if len(c.args) == 0 {
		return nil
	}
	return c.args[1:]
}
