package composition

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/console"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/filescore"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/osprocess"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/tcellui"
)

// RunnerCreator constructs the application.Runner and every outbound
// adapter it needs.
type RunnerCreator struct{}

// NewRunnerCreator returns a RunnerCreator.
func NewRunnerCreator() RunnerCreator {
	return RunnerCreator{}
}

// Create builds a Runner, along with every adapter it depends on.
func (RunnerCreator) Create() (inbound.Runner, error) {
	proc, err := osprocess.NewProcess()
	if err != nil {
		return nil, err
	}

	store, err := filescore.NewFileScore()
	if err != nil {
		return nil, err
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	ui := tcellui.NewTUI(screen)

	processSvc := process.NewService(proc, console.NewProcessReporter())
	scoreSvc := score.NewService(store, console.NewScoreReporter())
	gameSvc := game.NewService(processSvc, ui, ui.InputEvents())

	return application.NewRunner(processSvc, scoreSvc, gameSvc), nil
}
