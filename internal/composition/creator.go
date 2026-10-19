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
	"github.com/eirikur-ari/pidshooter/internal/entrypoint/cli"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/console"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/filestore"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/logger"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/pscommand"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
)

// Creator creates the program that runs the commands.
type Creator struct {
	errHandler *apperror.Handler
}

// NewCreator returns a Creator that logs errors to standard error.
func NewCreator() *Creator {
	return &Creator{errHandler: apperror.NewHandler(logger.NewLogger())}
}

// Create returns the program that runs the commands. Nothing a command needs is
// set up until that command runs.
func (c *Creator) Create() *cli.Program {
	return cli.NewProgram(
		&lazyRunner[inbound.PlayRequest, inbound.PlayRunner]{build: c.buildPlay, errHandler: c.errHandler},
		runner.NewList(c.errHandler),
		runner.NewKill(c.errHandler),
		c.errHandler,
	)
}

// buildPlay returns the runner of a game session wired with everything it
// depends on. It returns an error if any of those dependencies cannot be set up.
func (c *Creator) buildPlay() (inbound.PlayRunner, error) {
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

	return runner.NewPlay(configSvc, processSvc, scoreSvc, gameSvc, c.errHandler), nil
}
