// Package application wires together the services needed to run a game session
// and exposes that capability through the inbound.Runner port.
package application

import (
	"errors"
	"fmt"
	"os"

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
// them, then records and prints the resulting score. It returns an error
// whenever the caller needs to react to it: a fatal one (bad config,
// process discovery failure, or a failed game session) or a warning that
// still ends the run (no matching processes). Anything the run can
// continue past regardless (score load/save failures) is printed as it's
// encountered instead of being returned.
func (r *Runner) Run(cfg inbound.Config) error {
	if err := validateConfig(cfg); err != nil {
		return inbound.NewError(inbound.ErrorCodeInvalidConfig, inbound.ErrorSeverityFatal, "invalid configuration", err)
	}

	processes, err := r.processSvc.FindProcesses(cfg.Patterns)
	if err != nil {
		return inbound.NewError(inbound.ErrorCodeProcessDiscoveryFailed, inbound.ErrorSeverityFatal, "process discovery failed", err)
	}

	if len(processes) == 0 {
		msg := fmt.Sprintf("no processes found matching %v", cfg.Patterns)
		return inbound.NewError(inbound.ErrorCodeNoProcessesFound, inbound.ErrorSeverityWarning, msg, nil)
	}

	board, highScore, loadErr := r.scoreSvc.LoadScoreBoard()
	notFound := errors.As(loadErr, &outbound.NotFoundError{})
	if loadErr != nil && !notFound {
		inbound.NewError(inbound.ErrorCodeScoreLoadFailed, inbound.ErrorSeverityWarning, "could not load scores", loadErr).Fprint(os.Stderr)
	}

	result, err := r.gameSvc.Play(cfg, processes, highScore)
	if err != nil {
		return inbound.NewError(inbound.ErrorCodeGameFailed, inbound.ErrorSeverityFatal, "game session failed", err)
	}

	if recErr := r.scoreSvc.RecordScore(board, result.Kills, result.FreedMem, result.LowestSpeed, cfg.TimeLimit, result.Duration, loadErr == nil || notFound); recErr != nil {
		inbound.NewError(inbound.ErrorCodeScoreSaveFailed, inbound.ErrorSeverityWarning, "score not saved", recErr).Fprint(os.Stderr)
	}

	score.PrintResults(result.Duration, result.Kills, result.FreedMem, board)

	return nil
}
