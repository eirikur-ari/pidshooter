package composition

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/runner"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/console"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/filestore"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/logger"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/pscommand"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
)

// Runner runs the commands of pidshooter.
type Runner struct {
	errHandler *apperror.Handler
}

// NewRunner returns a Runner that logs errors to standard error.
func NewRunner() *Runner {
	return &Runner{errHandler: apperror.NewHandler(logger.NewLogger())}
}

// Play plays a game session as described by request. It returns an error if
// the session cannot be set up or if it fails.
func (r *Runner) Play(request inbound.PlayRequest) error {
	playRunner, err := r.buildPlay()
	if err != nil {
		return err
	}
	return playRunner.Run(request)
}

// HandleError logs err at the level its severity calls for. It returns nil if
// the error is absorbed, and err unchanged otherwise.
func (r *Runner) HandleError(err error) error {
	return r.errHandler.Handle(err)
}

// buildPlay returns the runner of a game session wired with everything it
// depends on. It returns an error if any of those dependencies cannot be set up.
func (r *Runner) buildPlay() (inbound.PlayRunner, error) {
	manager, err := pscommand.NewManager()
	if err != nil {
		return nil, err
	}

	scoreStore, err := filestore.NewScoreFile()
	if err != nil {
		return nil, err
	}

	configStore, err := filestore.NewConfigFile()
	if err != nil {
		return nil, err
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	ui := tcellui.NewTUI(screen)

	configSvc := config.NewService(configStore)
	processSvc := process.NewService(manager, console.NewProcessReporter())
	scoreSvc := score.NewService(scoreStore, console.NewScoreReporter())
	gameSvc := game.NewService(processSvc, ui, ui.InputEvents())

	return runner.NewPlay(configSvc, processSvc, scoreSvc, gameSvc, r.errHandler), nil
}
