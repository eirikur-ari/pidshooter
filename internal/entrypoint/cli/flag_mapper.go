package cli

import (
	"flag"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// flagMapper maps the confirm, speed, time, and include-root flags the
// caller actually passed to an inbound.RunRequest.
type flagMapper struct {
	flagSet     *flag.FlagSet
	confirm     *bool
	speed       *float64
	timeLimit   *int
	includeRoot *bool
	req         inbound.RunRequest
}

// newFlagMapper returns a flagMapper for flagSet's confirm, speed, time,
// and include-root flags.
func newFlagMapper(flagSet *flag.FlagSet) *flagMapper {
	return &flagMapper{
		flagSet:     flagSet,
		confirm:     flagSet.Bool("confirm", false, ""),
		speed:       flagSet.Float64("speed", 0, ""),
		timeLimit:   flagSet.Int("time", 0, ""),
		includeRoot: flagSet.Bool("include-root", false, ""),
	}
}

// toRunRequest returns the inbound.RunRequest for patterns, including only
// the flags the caller actually passed — an omitted flag stays nil, leaving
// it to the resolved defaults.
func (m *flagMapper) toRunRequest(patterns []string) inbound.RunRequest {
	m.req = inbound.RunRequest{Patterns: patterns}
	m.flagSet.Visit(m.visit)
	return m.req
}

// visit records the given flag's value onto the request being built, if
// the flag is one this flagMapper tracks.
func (m *flagMapper) visit(f *flag.Flag) {
	switch f.Name {
	case "confirm":
		m.req.Config.Game.ConfirmMode = m.confirm
	case "speed":
		m.req.Config.Game.Speed = m.speed
	case "time":
		m.req.Config.Game.TimeLimit = m.timeLimit
	case "include-root":
		m.req.Config.Process.IncludeRoot = m.includeRoot
	}
}
