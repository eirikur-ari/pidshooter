package composition

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
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
	reporter := console.NewScoreReporter()
	return application.NewRunner(proc, store, reporter, ui, ui), nil
}
