// Package application wires together the services needed to run a game session
// and exposes that capability through the inbound.Runner port.
package application

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
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
	errHandler apperror.Handler
}

// NewRunner constructs a Runner with all required outbound ports injected.
func NewRunner(
	processMgr outbound.ProcessManager,
	store outbound.ScoreStore,
	renderer outbound.Renderer,
	events outbound.InputSource,
	logger outbound.Logger,
) *Runner {
	processSvc := process.NewService(processMgr)
	return &Runner{
		processSvc: processSvc,
		gameSvc:    game.NewService(processSvc, renderer, events),
		scoreSvc:   score.NewService(store),
		errHandler: apperror.NewHandler(logger),
	}
}

// Run discovers processes matching cfg.Patterns, plays a game session against
// them, then records and prints the resulting score. Every failure is
// logged here, regardless of severity. Run returns nil unless the failure
// was Fatal, in which case it's returned too so the caller can terminate
// the program.
func (r *Runner) Run(cfg inbound.Config) error {
	if err := validateConfig(cfg); err != nil {
		return r.errHandler.Handle(apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", err))
	}

	processes, err := r.processSvc.FindProcesses(cfg.Patterns)
	if err != nil {
		return r.errHandler.Handle(err)
	}

	board, highScore, loadErr := r.scoreSvc.LoadScoreBoard()
	if err := r.errHandler.Handle(loadErr); err != nil {
		return err
	}

	result, err := r.gameSvc.Play(cfg, processes, highScore)
	if err != nil {
		return r.errHandler.Handle(err)
	}

	r.logKillFailures(result.KillFailures)

	recErr := r.scoreSvc.RecordScore(board, result.Kills, result.FreedMem, result.LowestSpeed, cfg.TimeLimit, result.Duration, loadErr)
	if err := r.errHandler.Handle(recErr); err != nil {
		return err
	}

	score.PrintResults(result.Duration, result.Kills, result.FreedMem, board)

	return nil
}

// logKillFailures reports each target the run could not kill.
func (r *Runner) logKillFailures(failures []game.KillFailure) {
	for _, f := range failures {
		msg := fmt.Sprintf("could not kill %s (PID %d)", f.Target, f.PID)
		_ = r.errHandler.Handle(apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, msg, f.Err))
	}
}
