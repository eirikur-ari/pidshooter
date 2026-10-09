package runner

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
)

// configLoader resolves the run configuration.
type configLoader interface {
	// Load returns the configuration for this run.
	Load() (config.Result, error)
}

// processFinder finds the processes that can be targeted.
type processFinder interface {
	// FindProcesses returns the processes to target for the given request.
	FindProcesses(request process.FindRequest) ([]process.FindResult, error)
}

// scoreKeeper loads, records and reports scores.
type scoreKeeper interface {
	// LoadScoreBoard returns the persisted score board, or an empty board
	// with an error if it cannot be loaded.
	LoadScoreBoard() (score.LoadResult, error)
	// RecordScore records a session on the request's score board. The result
	// is returned even when the error is not nil.
	RecordScore(request score.RecordRequest) (score.RecordResult, error)
	// ReportResults reports the outcome of a session against its score board.
	ReportResults(request score.ReportRequest)
}

// gamePlayer plays a game session.
type gamePlayer interface {
	// Play runs a game session over the request's processes.
	Play(request game.PlayRequest) (game.PlayResult, error)
}

// Service wires together the services required to run a game session. It implements inbound.Runner.
type Service struct {
	configSvc  configLoader
	processSvc processFinder
	gameSvc    gamePlayer
	scoreSvc   scoreKeeper
	errHandler *apperror.Handler
}

// NewService constructs a Service from its already-assembled collaborators.
func NewService(
	configSvc configLoader,
	processSvc processFinder,
	scoreSvc scoreKeeper,
	gameSvc gamePlayer,
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
	cfg, cfgErr := s.configSvc.Load()
	if err := s.errHandler.Handle(cfgErr); err != nil {
		return err
	}

	found, findErr := s.processSvc.FindProcesses(toFindRequest(cfg.Process))
	if findErr != nil {
		return s.errHandler.Handle(findErr)
	}

	loaded, loadErr := s.scoreSvc.LoadScoreBoard()
	if err := s.errHandler.Handle(loadErr); err != nil {
		return err
	}

	result, playErr := s.gameSvc.Play(toPlayRequest(cfg.Game, found, loaded.HighScore))
	if playErr != nil {
		return s.errHandler.Handle(playErr)
	}

	if err := s.errHandler.HandleAll(result.Errors); err != nil {
		return err
	}

	recorded, recordErr := s.scoreSvc.RecordScore(toRecordRequest(loaded, result, cfg.Game.TimeLimit))
	if err := s.errHandler.Handle(recordErr); err != nil {
		return err
	}

	s.scoreSvc.ReportResults(toReportRequest(result, recorded))

	return nil
}
