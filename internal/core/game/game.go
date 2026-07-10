// Package game implements the terminal-based PID shooter game logic.
package game

import (
	"time"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// Config holds the gameplay parameters for a session.
type Config struct {
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}

// Game manages targets and session state as a pure state machine.
// The application layer owns the loop, renderer, event source, and process killer.
type Game struct {
	cfg       Config
	state     state
	confirm   Confirmation
	velocity  Velocity
	processes []process.Info
	targets   []*Target
	Session
}

// New creates a new Game with the given configuration. Call Start before the first Update.
func New(processes []process.Info, cfg Config) *Game {
	return &Game{
		cfg:       cfg,
		processes: processes,
		targets:   make([]*Target, 0, len(processes)),
		velocity:  NewVelocity(cfg.Speed),
		confirm:   NewConfirmation(cfg.ConfirmMode),
	}
}

// Start transitions the game from Pending to Running. Panics if called on a
// game that is already running or stopped.
func (g *Game) Start(w, h int) {
	switch g.state.Load() {
	case Pending:
		g.initialize(w, h)
	case Running:
		panic("game: Start called on a running game")
	case Stopped:
		panic("game: Start called on a stopped game")
	}
}

// State returns the current lifecycle of the game.
func (g *Game) State() Lifecycle { return g.state.Load() }

// IsRunning reports whether the game loop should continue.
func (g *Game) IsRunning() bool { return g.state.Load() == Running }

// Stop transitions the game to Stopped, signaling the loop to exit.
func (g *Game) Stop() { g.state.Store(Stopped) }

func (g *Game) initialize(w, h int) {
	g.startTime = time.Now()
	for _, p := range g.processes {
		g.targets = append(g.targets, NewTarget(p, w, h))
	}
	g.state.Store(Running)
}
