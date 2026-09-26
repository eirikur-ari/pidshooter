package cli

import (
	"flag"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// toRunRequest maps parsed flags to an inbound.RunRequest, using
// flagSet.Visit to set a pointer only for flags the caller actually
// passed — an unpassed flag stays nil, leaving it to the resolved defaults.
func toRunRequest(flagSet *flag.FlagSet, patterns []string, input *programInput) inbound.RunRequest {
	req := inbound.RunRequest{Patterns: patterns}
	flagSet.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "confirm":
			req.Game.ConfirmMode = &input.confirm
		case "speed":
			req.Game.Speed = &input.speed
		case "time":
			req.Game.TimeLimit = &input.timeLimit
		case "include-root":
			req.Process.IncludeRoot = &input.includeRoot
		}
	})
	return req
}
