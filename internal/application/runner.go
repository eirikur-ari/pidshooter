// Package application wires together the services needed to run a game session
// and exposes that capability through the inbound.Runner port.
package application

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
)

// Runner wires together the services required to run a game session. It implements inbound.Runner.
type Runner struct {
	processSvc *process.Service
	gameSvc    *game.Service
	scoreSvc   *score.Service
}

// NewRunner constructs a Runner with all required outbound ports injected.
func NewRunner(
	processMgr outbound.ProcessManager,
	store outbound.ScoreStore,
	renderer outbound.Renderer,
	events outbound.InputSource,
) *Runner {
	processSvc := process.NewService(processMgr)
	return &Runner{
		processSvc: processSvc,
		gameSvc:    game.NewService(processSvc, renderer, events),
		scoreSvc:   score.NewService(store),
	}
}

// Run discovers processes matching cfg.Patterns, plays a game session against
// them, then records and prints the resulting score.
func (r *Runner) Run(cfg inbound.Config) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	processes, err := r.processSvc.FindProcesses(cfg.Patterns)
	if err != nil {
		return err
	}

	if len(processes) == 0 {
		return nil
	}

	board, highScore, success := r.scoreSvc.LoadScoreBoard()

	result, err := r.gameSvc.Play(cfg, processes, highScore)
	if err != nil {
		return err
	}

	r.scoreSvc.RecordScore(board, result.Kills, result.FreedMem, result.LowestSpeed, cfg.TimeLimit, result.Duration, success)

	score.PrintResults(result.Duration, result.Kills, result.FreedMem, board)

	return nil
}
