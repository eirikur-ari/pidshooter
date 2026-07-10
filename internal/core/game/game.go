// Package game implements the terminal-based PID shooter game logic.
package game

import (
	"sync/atomic"
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
	cfg        Config
	processes  []process.Info
	targets    []*Target
	velocity Velocity
	confirm  Confirmation
	running  atomic.Bool
	Session
}

// New creates a new Game with the given configuration. Call Init before the first Update.
func New(processes []process.Info, cfg Config) *Game {
	g := &Game{
		cfg:       cfg,
		processes: processes,
		targets:   make([]*Target, 0, len(processes)),
		velocity:  NewVelocity(cfg.Speed),
		confirm:   NewConfirmation(cfg.ConfirmMode),
	}
	g.running.Store(true)
	return g
}

// Init populates the game board and records the start time.
// Must be called once before the first Update.
func (g *Game) Init(w, h int) {
	g.startTime = time.Now()
	for _, p := range g.processes {
		g.targets = append(g.targets, NewTarget(p, w, h))
	}
}

// Running reports whether the game loop should continue.
func (g *Game) Running() bool { return g.running.Load() }

// Stop signals the game to exit on the next Running check.
func (g *Game) Stop() { g.running.Store(false) }
