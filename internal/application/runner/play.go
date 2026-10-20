package runner

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
)

// configLoader resolves the run configuration.
type configLoader interface {
	// Load returns the configuration for this run, with options taking
	// precedence over any stored configuration.
	Load(options config.Options) (config.Result, error)
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

// Play plays a game session over the processes that match search patterns.
type Play struct {
	configSvc  configLoader
	processSvc processFinder
	gameSvc    gamePlayer
	scoreSvc   scoreKeeper
	errHandler *apperror.Handler
}

// NewPlay constructs a Play from its already-assembled collaborators.
func NewPlay(
	configSvc configLoader,
	processSvc processFinder,
	scoreSvc scoreKeeper,
	gameSvc gamePlayer,
	errHandler *apperror.Handler,
) *Play {
	return &Play{
		configSvc:  configSvc,
		processSvc: processSvc,
		gameSvc:    gameSvc,
		scoreSvc:   scoreSvc,
		errHandler: errHandler,
	}
}

// Run resolves the final run configuration from the request, discovers the
// processes matching its patterns, plays a game session against them, then
// records and prints the resulting score. Every failure is logged, regardless
// of severity. Run returns nil unless the failure was Fatal, in which case the
// error is returned.
func (p *Play) Run(request inbound.PlayRequest) error {
	cfg, cfgErr := p.configSvc.Load(toConfigOptions(request))
	if err := p.errHandler.Handle(cfgErr); err != nil {
		return err
	}

	found, findErr := p.processSvc.FindProcesses(toFindRequest(request.Patterns, cfg.Process))
	if findErr != nil {
		return p.errHandler.Handle(findErr)
	}

	loaded, loadErr := p.scoreSvc.LoadScoreBoard()
	if err := p.errHandler.Handle(loadErr); err != nil {
		return err
	}

	result, playErr := p.gameSvc.Play(toPlayRequest(cfg.Game, found, loaded.HighScore))
	if playErr != nil {
		return p.errHandler.Handle(playErr)
	}

	if err := p.errHandler.HandleAll(result.Errors); err != nil {
		return err
	}

	recorded, recordErr := p.scoreSvc.RecordScore(toRecordRequest(loaded, result, cfg.Game.TimeLimit))
	if err := p.errHandler.Handle(recordErr); err != nil {
		return err
	}

	p.scoreSvc.ReportResults(toReportRequest(result, recorded))

	return nil
}
