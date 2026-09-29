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
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/osprocess"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
)

// RunnerCreator constructs a Runner, along with everything it depends on.
type RunnerCreator struct {
	errHandler *apperror.Handler
}

// NewRunnerCreator returns a RunnerCreator.
func NewRunnerCreator() RunnerCreator {
	return RunnerCreator{errHandler: apperror.NewHandler(logger.NewLogger())}
}

// ErrHandler returns the Handler used to log errors.
func (c RunnerCreator) ErrHandler() *apperror.Handler {
	return c.errHandler
}

// Create builds a Runner, along with every adapter it depends on.
func (c RunnerCreator) Create() (inbound.Runner, error) {
	proc, err := osprocess.NewProcess()
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
	processSvc := process.NewService(proc, console.NewProcessReporter())
	scoreSvc := score.NewService(scoreStore, console.NewScoreReporter())
	gameSvc := game.NewService(processSvc, ui, ui.InputEvents())

	return runner.NewService(configSvc, processSvc, scoreSvc, gameSvc, c.errHandler), nil
}
