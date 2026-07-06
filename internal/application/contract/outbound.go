package contract

import (
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

// Renderer is the outbound port for terminal rendering.
type Renderer interface {
	Init() error
	Cleanup()
	Size() (width, height int)
	Render(frame game.Frame)
}

// InputEvent is the application port type for all user input events.
type InputEvent interface{}

// EventSource is the outbound port for user input events.
type EventSource interface {
	Events() <-chan InputEvent
}

// ProcessKiller is the outbound port for sending a kill signal to a process.
// name is the expected process name as discovered at game start; Kill must
// verify the process still has that name before sending the signal.
type ProcessKiller interface {
	Kill(pid int, name string) error
}

// Finder is the outbound port for process discovery on the host.
type Finder interface {
	List() ([]process.Info, error)
	Find(patterns []string) ([]process.Info, error)
}

// Store is the outbound port for persisting and retrieving the score board.
type Store interface {
	Load() (*score.Board, error)
	Save(board *score.Board) error
}
