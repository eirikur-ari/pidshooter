// Package application wires together the services needed to run a game session
// and exposes that capability through the inbound.Runner port.
package application

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
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
	proc outbound.Process,
	store outbound.ScoreStore,
	reporter outbound.ScoreReporter,
	renderer outbound.Renderer,
	events outbound.InputEventProvider,
) *Runner {
	processSvc := process.NewService(proc)
	return &Runner{
		processSvc: processSvc,
		gameSvc:    game.NewService(processSvc, renderer, events),
		scoreSvc:   score.NewService(store, reporter),
	}
}

// Run discovers processes matching cfg.Patterns, plays a game session against
// them, then records and prints the resulting score. Every failure is
// logged here, regardless of severity. Run returns nil unless the failure
// was Fatal, in which case it's returned too so the caller can terminate
// the program.
func (r *Runner) Run(cfg config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	processes, err := r.processSvc.FindProcesses(cfg.Patterns, cfg.IncludeRoot)
	if err != nil {
		return apperror.Handle(err)
	}

	board, highScore, loadErr := r.scoreSvc.LoadScoreBoard()
	if err := apperror.Handle(loadErr); err != nil {
		return err
	}

	result, err := r.gameSvc.Play(cfg, processes, highScore)
	if err != nil {
		return apperror.Handle(err)
	}

	r.logKillFailures(result.KillFailures)
	r.logDuds(result.Duds)

	entry := score.ToEntry(result, cfg.TimeLimit)
	recErr := r.scoreSvc.RecordScore(board, entry, loadErr)
	if err := apperror.Handle(recErr); err != nil {
		return err
	}

	r.scoreSvc.ReportResults(result.Duration, result.Kills, len(result.Duds), result.FreedMem, board)

	return nil
}

// logKillFailures reports each target the run could not kill.
func (r *Runner) logKillFailures(failures []game.KillFailure) {
	for _, f := range failures {
		msg := fmt.Sprintf("could not kill %s (PID %d)", f.Target, f.PID)
		_ = apperror.Handle(apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, msg, f.Err))
	}
}

// logDuds reports each target whose backing process was already gone
// before a kill could land on it.
func (r *Runner) logDuds(duds []game.KillDud) {
	for _, d := range duds {
		msg := fmt.Sprintf("%s (PID %d) ran away before it could be killed", d.Target, d.PID)
		_ = apperror.Handle(apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, msg, nil))
	}
}
