package cli

import (
	"flag"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// flagMapper maps the flags actually passed to config options.
type flagMapper struct {
	flagSet     *flag.FlagSet
	confirm     *bool
	speed       *float64
	timeLimit   *int
	includeRoot *bool
	allowRoot   *bool
}

// newFlagMapper returns a flagMapper that defines its flags on flagSet.
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

// toConfigOptions returns the config options built from the flags actually
// passed — an omitted flag stays nil.
func (m *flagMapper) toConfigOptions() config.Options {
	var options config.Options
	m.flagSet.Visit(func(passedFlag *flag.Flag) {
		switch passedFlag.Name {
		case "confirm":
			options.Game.ConfirmMode = util.ClonePointer(m.confirm)
		case "speed":
			options.Game.Speed = util.ClonePointer(m.speed)
		case "time":
			options.Game.TimeLimit = util.ClonePointer(m.timeLimit)
		case "include-root":
			options.Process.IncludeRoot = util.ClonePointer(m.includeRoot)
		case "i-am-root":
			options.Process.AllowRoot = util.ClonePointer(m.allowRoot)
		}
	})
	return options
}
