package cli

import (
	"flag"
	"fmt"
	"strings"
)

// flagSplitter splits command-line arguments into search patterns and flag
// tokens, using flagSet's registered flags to tell them apart.
type flagSplitter struct {
	flagSet *flag.FlagSet
}

// newFlagSplitter returns a flagSplitter backed by flagSet.
func newFlagSplitter(flagSet *flag.FlagSet) flagSplitter {
	return flagSplitter{flagSet: flagSet}
}

// split partitions args into search patterns and flag tokens. Any argument
// that isn't a flag registered on flagSet (or that flag's value) is
// treated as a pattern, regardless of where it falls among the flags. A
// "--" argument ends flag processing; every argument after it is treated
// as a pattern unconditionally. split classifies every token this way, so
// the flagSet it feeds never has a leftover positional to stop on.
func (p flagSplitter) split(args []string) (patterns, flagArgs []string, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			patterns = append(patterns, args[i+1:]...)
			break
		}

		if arg == "-" || !strings.HasPrefix(arg, "-") {
			patterns = append(patterns, arg)
			continue
		}

		name, hasValue := flagNameAndValue(arg)
		flagState := p.flagSet.Lookup(name)

		if flagState == nil {
			return nil, nil, fmt.Errorf("flag provided but not defined: -%s", name)
		}

		flagArgs = append(flagArgs, arg)

		if hasValue {
			continue
		}

		if isBoolFlag(flagState) {
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

// looksLikeFlag reports whether arg would itself be classified as a flag
// token by split, rather than consumed as a bare value.
func looksLikeFlag(arg string) bool {
	return arg != "-" && strings.HasPrefix(arg, "-")
}

// isBoolFlag reports whether flagState is a boolean flag, per the flag
// package's own convention: a boolean flag's Value optionally implements
// IsBoolFlag() bool.
func isBoolFlag(flagState *flag.Flag) bool {
	boolValue, ok := flagState.Value.(interface{ IsBoolFlag() bool })
	return ok && boolValue.IsBoolFlag()
}

// flagNameAndValue extracts a flag's name from a "-name" or "--name=value"
// token, reporting whether the value was inlined with "=".
func flagNameAndValue(arg string) (name string, hasValue bool) {
	name = strings.TrimPrefix(strings.TrimPrefix(arg, "-"), "-")

	if flagName, _, found := strings.Cut(name, "="); found {
		return flagName, true
	}

	return name, false
}
