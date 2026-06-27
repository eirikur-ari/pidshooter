// Package game implements the terminal-based PID shooter game loop.
package game

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	gamedriven "github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"
	procdriven "github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"
)

// Game manages targets, session state, and user interaction.
type Game struct {
	renderer    gamedriven.Renderer
	events      gamedriven.EventSource
	killer      gamedriven.ProcessKiller
	processes   []procdriven.Info
	targets     []*Target
	confirmMode bool
	speed       float64
	timeLimit   int
	running     atomic.Bool
	confirming  *Target
	Session
}

// New creates a new game instance with the given configuration and driven ports.
func New(
	processes []procdriven.Info,
	confirmMode bool,
	speed float64,
	timeLimit int,
	killer gamedriven.ProcessKiller,
	renderer gamedriven.Renderer,
	events gamedriven.EventSource,
) *Game {
	g := &Game{
		renderer:    renderer,
		events:      events,
		killer:      killer,
		processes:   processes,
		targets:     make([]*Target, 0, len(processes)),
		confirmMode: confirmMode,
		speed:       speed,
		timeLimit:   timeLimit,
	}
	g.running.Store(true)
	return g
}

// Play runs the full game lifecycle: init renderer, run loop, cleanup.
func (g *Game) Play(highScore int) error {
	g.SetHighScore(highScore)

	if err := g.renderer.Init(); err != nil {
		return fmt.Errorf("renderer initialization failed: %w", err)
	}
	defer g.renderer.Cleanup()

	w, h := g.renderer.Size()
	g.populateEntities(g.processes, w, h)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	defer signal.Stop(sigCh)
	go func() {
		<-sigCh
		g.stop()
	}()

	g.run()
	return nil
}

func (g *Game) stop() { g.running.Store(false) }
