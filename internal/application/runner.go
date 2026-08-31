// Package application wires together the services needed to run a game session
// and exposes that capability through the inbound.Runner port.
package application

import (
	"errors"

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
	logger     outbound.Logger
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
		logger:     logger,
	}
}

// Run discovers processes matching cfg.Patterns, plays a game session against
// them, then records and prints the resulting score. It returns an error
// whenever the caller needs to react to it: a fatal one (bad config,
// process discovery failure, or a failed game session) or a warning that
// still ends the run (no matching processes). Anything the run can
// continue past regardless — score load/save failures, and targets the
// session could not kill — is reported through the logger port instead of
// being returned.
func (r *Runner) Run(cfg inbound.Config) error {
	if err := validateConfig(cfg); err != nil {
		return apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", err)
	}

	processes, err := r.processSvc.FindProcesses(cfg.Patterns)
	if err != nil {
		return err
	}

	board, highScore, loadErr := r.scoreSvc.LoadScoreBoard()
	notFound := errors.As(loadErr, &outbound.NotFoundError{})
	if loadErr != nil && !notFound {
		r.logWarning(apperror.CodeScoreLoadFailed, "could not load scores", loadErr)
	}

	result, err := r.gameSvc.Play(cfg, processes, highScore)
	if err != nil {
		return err
	}

	r.logKillFailures(result.KillFailureMessages)

	if recErr := r.scoreSvc.RecordScore(board, result.Kills, result.FreedMem, result.LowestSpeed, cfg.TimeLimit, result.Duration, loadErr == nil || notFound); recErr != nil {
		r.logWarning(apperror.CodeScoreSaveFailed, "score not saved", recErr)
	}

	score.PrintResults(result.Duration, result.Kills, result.FreedMem, board)

	return nil
}

// logKillFailures reports each target the run could not kill.
func (r *Runner) logKillFailures(messages []string) {
	for _, msg := range messages {
		r.logWarning(apperror.CodeKillFailed, msg, nil)
	}
}

// logWarning reports a failure the run continued past, formatted identically
// to a Warning-severity apperror.Error returned to the caller.
func (r *Runner) logWarning(code apperror.Code, message string, err error) {
	r.logger.Warn(apperror.NewError(code, apperror.SeverityWarning, message, err).Error())
}
