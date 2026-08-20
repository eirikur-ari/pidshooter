// Package application wires together the services needed to run a game session
// and exposes that capability through the inbound.Runner port.
package application

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
)

// Runner wires together the services required to run a game session. It implements inbound.Runner.
type Runner struct {
	game *game.Service
}

// NewRunner constructs a Runner with all required outbound ports injected.
func NewRunner(
	process outbound.Process,
	store outbound.ScoreStore,
	renderer outbound.Renderer,
	events outbound.InputSource,
) *Runner {
	return &Runner{game: game.NewService(process, store, renderer, events)}
}

// Run starts a game session with the given configuration.
func (r *Runner) Run(cfg inbound.Config) error {
	return r.game.Play(cfg)
}
