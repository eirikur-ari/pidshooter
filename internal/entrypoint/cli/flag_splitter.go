package cli

import (
	"flag"
	"fmt"
	"strings"
)

// flagSplitter splits command-line arguments into search patterns and flag
// arguments, telling them apart by the flags registered on flagSet.
type flagSplitter struct {
	flagSet *flag.FlagSet
}

// newFlagSplitter returns a flagSplitter that recognizes the flags registered
// on flagSet.
func newFlagSplitter(flagSet *flag.FlagSet) flagSplitter {
	return flagSplitter{flagSet: flagSet}
}

// split partitions the given arguments into search patterns and flag
// arguments. An argument that isn't a registered flag, or the value of one, is
// a pattern wherever it falls among the flags. A "--" argument ends flag
// processing: every argument after it is a pattern. It returns an error for an
// unregistered flag, or for a flag that needs a value and has none.
func (s flagSplitter) split(args []string) (patterns, flagArgs []string, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			patterns = append(patterns, args[i+1:]...)
			break
		}

		if !looksLikeFlag(arg) {
			patterns = append(patterns, arg)
			continue
		}

		name, hasValue := flagNameAndValue(arg)
		registeredFlag := s.flagSet.Lookup(name)

		if registeredFlag == nil {
			return nil, nil, fmt.Errorf("flag provided but not defined: -%s", name)
		}

		flagArgs = append(flagArgs, arg)

		if hasValue {
			continue
		}

		if isBoolFlag(registeredFlag) {
			continue
		}

		if i+1 >= len(args) || looksLikeFlag(args[i+1]) {
			return nil, nil, fmt.Errorf("flag needs an argument: -%s", name)
		}

		i++
		flagArgs = append(flagArgs, args[i])
	}

	return patterns, flagArgs, nil
}

// looksLikeFlag reports whether the argument starts with "-" and is not a lone
// "-".
func looksLikeFlag(arg string) bool {
	return arg != "-" && strings.HasPrefix(arg, "-")
}

// isBoolFlag reports whether the registered flag is a boolean flag.
func isBoolFlag(registeredFlag *flag.Flag) bool {
	boolValue, ok := registeredFlag.Value.(interface{ IsBoolFlag() bool })
	return ok && boolValue.IsBoolFlag()
}

// flagNameAndValue extracts a flag's name from a "-name" or "--name=value"
// argument, reporting whether the value was inlined with "=".
func flagNameAndValue(arg string) (name string, hasValue bool) {
	name = strings.TrimPrefix(strings.TrimPrefix(arg, "-"), "-")

	if flagName, _, found := strings.Cut(name, "="); found {
		return flagName, true
	}

	return name, false
}
