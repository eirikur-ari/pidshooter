package cli

import (
	"flag"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// flagMapper maps the confirm, speed, time, include-root, and i-am-root
// flags the caller actually passed to a config.Options.
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

// toConfigOptions returns the config.Options built from the flags the
// caller actually passed — an omitted flag stays nil, leaving it to the
// resolved defaults.
func (m *flagMapper) toConfigOptions() config.Options {
	var opts config.Options
	m.flagSet.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "confirm":
			opts.Game.ConfirmMode = util.ClonePointer(m.confirm)
		case "speed":
			opts.Game.Speed = util.ClonePointer(m.speed)
		case "time":
			opts.Game.TimeLimit = util.ClonePointer(m.timeLimit)
		case "include-root":
			opts.Process.IncludeRoot = util.ClonePointer(m.includeRoot)
		case "i-am-root":
			opts.Process.AllowRoot = util.ClonePointer(m.allowRoot)
		}
	})
	return opts
}
