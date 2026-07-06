// Package game implements the terminal-based PID shooter game logic.
package game

import (
	"sync/atomic"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// Game manages targets and session state as a pure state machine.
// The application layer owns the loop, renderer, event source, and process killer.
type Game struct {
	processes   []process.Info
	targets     []*Target
	confirmMode bool
	speed       float64
	timeLimit   int
	running     atomic.Bool
	confirming  *Target
	Session
}

// KillRequest is returned by HandleClick or HandleKey when a target should be killed.
// The application layer performs the actual kill and calls CompleteKill on success.
type KillRequest struct {
	Target *Target
}

// New creates a new Game with the given configuration. Call Init before the first Update.
func New(processes []process.Info, confirmMode bool, speed float64, timeLimit int) *Game {
	g := &Game{
		processes:   processes,
		targets:     make([]*Target, 0, len(processes)),
		confirmMode: confirmMode,
		speed:       speed,
		timeLimit:   timeLimit,
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
