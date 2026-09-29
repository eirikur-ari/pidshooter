package config

import (
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
)

// GameRequest holds the caller's explicitly-provided game mode values. A
// nil field means the caller did not provide that value.
type GameRequest struct {
	ConfirmMode *bool
	Speed       *float64
	TimeLimit   *int
}

// ProcessRequest holds the caller's explicitly-provided process-discovery
// values. A nil field means the caller did not provide that value.
type ProcessRequest struct {
	IncludeRoot *bool
	// AllowRoot permits running pidshooter itself as root; never read from
	// the config file, only ever set per invocation.
	AllowRoot *bool
}

// Request holds the caller's explicitly-provided values.
type Request struct {
	Game    GameRequest
	Process ProcessRequest
}

// validate returns an error if any explicitly-provided value fails
// validation. Fields left unset are not validated.
func (r Request) validate() error {
	if r.Game.Speed != nil {
		if err := movement.ValidateSpeed(*r.Game.Speed); err != nil {
			return err
		}
	}
	if r.Game.TimeLimit != nil {
		if err := game.ValidateTimeLimit(*r.Game.TimeLimit); err != nil {
			return err
		}
	}
	return nil
}
