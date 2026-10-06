package runner

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
)

// Service wires together the services required to run a game session. It implements inbound.Runner.
type Service struct {
	configSvc  *config.Service
	processSvc *process.Service
	gameSvc    *game.Service
	scoreSvc   *score.Service
	errHandler *apperror.Handler
}

// NewService constructs a Service from its already-assembled collaborators.
func NewService(
	configSvc *config.Service,
	processSvc *process.Service,
	scoreSvc *score.Service,
	gameSvc *game.Service,
	errHandler *apperror.Handler,
) *Service {
	return &Service{
		configSvc:  configSvc,
		processSvc: processSvc,
		gameSvc:    gameSvc,
		scoreSvc:   scoreSvc,
		errHandler: errHandler,
	}
}

// Run resolves the final run configuration, discovers matching processes,
// plays a game session against them, then records and prints the
// resulting score. Every failure is logged here, regardless of severity.
// Run returns nil unless the failure was Fatal, in which case it's
// returned too so the caller can terminate the program.
func (s *Service) Run() error {
	cfg, loadErr := s.configSvc.Load()
	if err := s.errHandler.Handle(loadErr); err != nil {
		return err
	}

	processes, err := s.processSvc.FindProcesses(toFindRequest(cfg.Process))
	if err != nil {
		return s.errHandler.Handle(err)
	}

	board, highScore, scoreLoadErr := s.scoreSvc.LoadScoreBoard()
	if err := s.errHandler.Handle(scoreLoadErr); err != nil {
		return err
	}

	result, err := s.gameSvc.Play(toPlayRequest(cfg.Game), processes, highScore)
	if err != nil {
		return s.errHandler.Handle(err)
	}

	s.logKillFailures(result.KillFailures)
	s.logDuds(result.Duds)

	entry := score.ToEntry(result, cfg.Game.TimeLimit)
	recErr := s.scoreSvc.RecordScore(board, entry, scoreLoadErr)
	if err := s.errHandler.Handle(recErr); err != nil {
		return err
	}

	s.scoreSvc.ReportResults(result.Duration, result.Kills, len(result.Duds), result.FreedMem, board)

	return nil
}

// logKillFailures reports each target the run could not kill.
func (s *Service) logKillFailures(failures []game.KillFailure) {
	for _, f := range failures {
		msg := fmt.Sprintf("could not kill %s (PID %d)", f.Name, f.PID)
		_ = s.errHandler.Handle(apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, msg, f.Err))
	}
}

// logDuds reports each target whose backing process was already gone
// before a kill could land on it.
func (s *Service) logDuds(duds []game.KillDud) {
	for _, d := range duds {
		msg := fmt.Sprintf("%s (PID %d) ran away before it could be killed", d.Name, d.PID)
		_ = s.errHandler.Handle(apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, msg, nil))
	}
}
