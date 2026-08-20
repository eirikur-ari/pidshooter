// Package application wires together the services needed to run a game session
// and exposes that capability through the inbound.Runner port.
package application

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
)

// Runner wires together the services required to run a game session. It implements inbound.Runner.
type Runner struct {
	game  *game.Service
	score *score.Service
}

// NewRunner constructs a Runner with all required outbound ports injected.
func NewRunner(
	process outbound.Process,
	store outbound.ScoreStore,
	renderer outbound.Renderer,
	events outbound.InputSource,
) *Runner {
	return &Runner{
		game:  game.NewService(process, renderer, events),
		score: score.NewService(store),
	}
}

// Run discovers processes matching cfg.Patterns, plays a game session against
// them, then records and prints the resulting score.
func (r *Runner) Run(cfg inbound.Config) error {
	processes, err := r.game.FindProcesses(cfg.Patterns)
	if err != nil {
		return err
	}

	if len(processes) == 0 {
		return nil
	}

	board, tracker, success := r.score.LoadScoreBoard()

	result, err := r.game.Play(cfg, processes, tracker)
	if err != nil {
		return err
	}

	r.score.RecordScore(board, result.LowestSpeed, cfg.TimeLimit, result.Duration, success)

	score.PrintResults(result.Duration, board)

	return nil
}
