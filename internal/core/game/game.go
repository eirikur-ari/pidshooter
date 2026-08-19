// Package game implements the terminal-based PID shooter game logic.
package game

import (
	"time"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
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
	cfg      Config
	state    atomicLifecycle
	timer    timer
	confirm  confirmation
	throttle *movement.Throttle
	roster   roster
}

// New creates a new Game with the given configuration. Call Start before the first Step.
func New(processes []process.Info, cfg Config) *Game {
	return &Game{
		cfg:      cfg,
		roster:   newRoster(processes),
		state:    newAtomicLifecycle(),
		timer:    newTimer(cfg.TimeLimit),
		throttle: movement.NewThrottle(cfg.Speed),
		confirm:  newConfirmation(cfg.Confirm),
	}
}

// Start transitions the game from pending to running. Panics if called on a
// game that is already running or stopped.
func (g *Game) Start(w, h int) {
	switch g.state.Load() {
	case pending:
		g.initialize(w, h)
	case running:
		panic("game: Start called on a running game")
	case stopped:
		panic("game: Start called on a stopped game")
	}
}

// Step advances the game state by one step. w and h are the current terminal dimensions.
// It moves every target and stops the game if the time limit has expired or all targets are dead.
func (g *Game) Step(w, h int) {
	if g.timer.Expired() {
		g.Stop()
		return
	}

	g.roster.move(movement.NewBounds(w, h), g.throttle.Speed())

	if g.roster.allDead() {
		g.Stop()
	}
}

// IsRunning reports whether the game loop should continue.
func (g *Game) IsRunning() bool { return g.state.Load() == running }

// Stop transitions the game to stopped, signaling the loop to exit.
func (g *Game) Stop() { g.state.Store(stopped) }

// StartTime returns when the game session was started.
func (g *Game) StartTime() time.Time { return g.timer.StartTime() }

// Throttle returns the game's throttle.
func (g *Game) Throttle() *movement.Throttle { return g.throttle }

// Targets returns the live target slice for frame assembly. Callers must not modify it.
func (g *Game) Targets() []*Target { return g.roster.targets }

// AvailableTargets returns the targets still in play — dead targets excluded —
// along with how many are currently alive.
func (g *Game) AvailableTargets() ([]*Target, int) { return g.roster.available() }

// TimeLimit returns the configured time limit in seconds (0 = unlimited).
func (g *Game) TimeLimit() int { return g.timer.LimitSeconds() }

// TimeLeft returns the remaining time in seconds.
func (g *Game) TimeLeft() int { return g.timer.SecondsLeft() }

// PendingConfirm returns the pending kill target, or nil if none is pending.
func (g *Game) PendingConfirm() *Target {
	if !g.confirm.Pending() {
		return nil
	}
	return g.confirm.target
}

// RequestConfirm registers target as pending confirmation (confirm mode) or returns
// it immediately for killing (passthrough mode).
func (g *Game) RequestConfirm(t *Target) *Target { return g.confirm.Request(t) }

func (g *Game) initialize(w, h int) {
	g.timer.Start()
	g.roster.spawn(movement.NewBounds(w, h))
	g.state.Store(running)
}

// currentState returns the current lifecycle of the game.
func (g *Game) currentState() lifecycle { return g.state.Load() }
