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
// them, then records and prints the resulting score. Every failure is
// logged here, regardless of severity. Run returns nil unless the failure
// was Fatal, in which case it's returned too so the caller can terminate
// the program.
func (r *Runner) Run(cfg inbound.Config) error {
	if err := validateConfig(cfg); err != nil {
		return r.handle(apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", err))
	}

	processes, err := r.processSvc.FindProcesses(cfg.Patterns)
	if err != nil {
		return r.handle(err)
	}

	board, highScore, loadErr := r.scoreSvc.LoadScoreBoard()
	notFound := errors.As(loadErr, &outbound.NotFoundError{})
	if loadErr != nil && !notFound {
		_ = r.handle(apperror.NewError(apperror.CodeScoreLoadFailed, apperror.SeverityWarning, "could not load scores", loadErr))
	}

	result, err := r.gameSvc.Play(cfg, processes, highScore)
	if err != nil {
		return r.handle(err)
	}

	r.logKillFailures(result.KillFailureMessages)

	if recErr := r.scoreSvc.RecordScore(board, result.Kills, result.FreedMem, result.LowestSpeed, cfg.TimeLimit, result.Duration, loadErr == nil || notFound); recErr != nil {
		_ = r.handle(apperror.NewError(apperror.CodeScoreSaveFailed, apperror.SeverityWarning, "score not saved", recErr))
	}

	score.PrintResults(result.Duration, result.Kills, result.FreedMem, board)

	return nil
}

// logKillFailures reports each target the run could not kill.
func (r *Runner) logKillFailures(messages []string) {
	for _, msg := range messages {
		_ = r.handle(apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, msg, nil))
	}
}

// handle logs err at the level its Severity calls for, then reports whether
// the caller must still treat the run as failed. A Warning-severity error
// is absorbed here, so Run reports success. A Fatal-severity error is
// returned after being logged, so the caller can terminate the program.
func (r *Runner) handle(err error) error {
	var appErr *apperror.Error
	errors.As(err, &appErr)
	if appErr.Severity == apperror.SeverityWarning {
		r.logger.Warn(err.Error())
		return nil
	}
	r.logger.Error(err.Error())
	if appErr.Severity == apperror.SeverityFatal {
		return err
	}
	return nil
}
