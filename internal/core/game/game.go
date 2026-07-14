// Package game implements the terminal-based PID shooter game logic.
package game

import (
	"time"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// Config holds the gameplay parameters for a session.
type Config struct {
	Confirm   bool
	Speed     float64
	TimeLimit int
}

// Game manages targets and session state as a pure state machine.
// The application layer owns the loop, renderer, event source, and process killer.
type Game struct {
	cfg       Config
	state     state
	timer     Timer
	confirm   Confirmation
	velocity  Velocity
	processes []process.Info
	targets   []*Target
	Stats
}

// New creates a new Game with the given configuration. Call Start before the first Update.
func New(processes []process.Info, cfg Config) *Game {
	return &Game{
		cfg:       cfg,
		processes: processes,
		targets:   make([]*Target, 0, len(processes)),
		timer:    NewTimer(cfg.TimeLimit),
		velocity: NewVelocity(cfg.Speed),
		confirm:  NewConfirmation(cfg.Confirm),
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

// StartTime returns when the game session was started.
func (g *Game) StartTime() time.Time { return g.timer.StartTime() }

// Targets returns the live target slice for frame assembly. Callers must not modify it.
func (g *Game) Targets() []*Target { return g.targets }

// Speed returns the current velocity speed.
func (g *Game) Speed() float64 { return g.velocity.Speed() }

// Velocity returns the game's velocity.
func (g *Game) Velocity() Speeder { return &g.velocity }

// TimeLimit returns the configured time limit in seconds (0 = unlimited).
func (g *Game) TimeLimit() int { return g.timer.LimitSeconds() }

// TimeLeft returns the remaining time in seconds.
func (g *Game) TimeLeft() int { return g.timer.SecondsLeft() }

// ConfirmTarget returns the pending kill target, or nil if none is pending.
func (g *Game) ConfirmTarget() *Target {
	if !g.confirm.Pending() {
		return nil
	}
	return g.confirm.target
}

// Confirm returns the game's confirmation state.
func (g *Game) Confirm() Confirmer { return &g.confirm }

func (g *Game) initialize(w, h int) {
	g.timer.Start()
	for _, p := range g.processes {
		g.targets = append(g.targets, NewTarget(p, FrameBounds{Width: w, Height: h}))
	}
	g.state.Store(Running)
}
