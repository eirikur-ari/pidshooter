// Package game implements the terminal-based PID shooter game loop.
package game

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/eirikur-ari/pidshooter/internal/process"
	"github.com/gdamore/tcell/v2"
)

// Game manages the terminal display, targets, and user interaction.
type Game struct {
	screen      tcell.Screen
	processes   []process.Info
	targets    []*Target
	confirmMode bool
	speed       float64
	timeLimit   int
	running     bool
	confirming  *Target
	Session
}

// New creates a new game instance.
func New(processes []process.Info, confirmMode bool, speed float64, timeLimit int) *Game {
	return &Game{
		processes:   processes,
		targets:    make([]*Target, 0, len(processes)),
		confirmMode: confirmMode,
		speed:       speed,
		running:     true,
		timeLimit:   timeLimit,
	}
}

// UseScreen injects a pre-created screen, bypassing tcell.NewScreen in init.
// Intended for tests using tcell.NewSimulationScreen.
func (g *Game) UseScreen(s tcell.Screen) {
	g.screen = s
}

// Play runs the full game lifecycle: init screen, run loop, cleanup.
func (g *Game) Play(highScore int) error {
	g.SetHighScore(highScore)

	if err := g.init(); err != nil {
		return fmt.Errorf("screen initialization failed: %w", err)
	}
	defer g.cleanup()

	g.populateEntities(g.processes)

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

func (g *Game) stop() { g.running = false }