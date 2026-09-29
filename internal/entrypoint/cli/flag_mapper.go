package cli

import (
	"flag"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// flagMapper maps the confirm, speed, time, include-root, and i-am-root
// flags the caller actually passed to an inbound.RunRequest.
type flagMapper struct {
	flagSet     *flag.FlagSet
	confirm     *bool
	speed       *float64
	timeLimit   *int
	includeRoot *bool
	allowRoot   *bool
}

// newFlagMapper returns a flagMapper for flagSet's confirm, speed, time,
// include-root, and i-am-root flags.
func newFlagMapper(flagSet *flag.FlagSet) *flagMapper {
	return &flagMapper{
		flagSet:     flagSet,
		confirm:     flagSet.Bool("confirm", false, ""),
		speed:       flagSet.Float64("speed", 0, ""),
		timeLimit:   flagSet.Int("time", 0, ""),
		includeRoot: flagSet.Bool("include-root", false, ""),
		allowRoot:   flagSet.Bool("i-am-root", false, ""),
	}
}

// toRunRequest returns the inbound.RunRequest for patterns, including only
// the flags the caller actually passed — an omitted flag stays nil, leaving
// it to the resolved defaults.
func (m *flagMapper) toRunRequest(patterns []string) inbound.RunRequest {
	req := inbound.RunRequest{Patterns: patterns}
	m.flagSet.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "confirm":
			req.Config.Game.ConfirmMode = util.ClonePtr(m.confirm)
		case "speed":
			req.Config.Game.Speed = util.ClonePtr(m.speed)
		case "time":
			req.Config.Game.TimeLimit = util.ClonePtr(m.timeLimit)
		case "include-root":
			req.Config.Process.IncludeRoot = util.ClonePtr(m.includeRoot)
		case "i-am-root":
			req.Config.Process.AllowRoot = util.ClonePtr(m.allowRoot)
		}
	})
	return req
}
