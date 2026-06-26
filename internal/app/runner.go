// Package app contains the application service: use-case orchestration.
package app

import (
	"fmt"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/domain/game"
	gamedriven "github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"
	"github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driving"
	procdriven "github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"
	"github.com/eirikur-ari/pidshooter/internal/domain/score"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// GameService implements driving.GameServicePort by orchestrating domain objects and driven ports.
type GameService struct {
	finder   procdriven.Finder
	killer   gamedriven.ProcessKiller
	store    score.Store
	renderer gamedriven.Renderer
	events   gamedriven.EventSource
}

// NewGameService constructs a GameService with all required driven ports injected.
func NewGameService(
	finder procdriven.Finder,
	killer gamedriven.ProcessKiller,
	store score.Store,
	renderer gamedriven.Renderer,
	events gamedriven.EventSource,
) *GameService {
	return &GameService{
		finder:   finder,
		killer:   killer,
		store:    store,
		renderer: renderer,
		events:   events,
	}
}

// Play runs a complete game session: discovery → game → score persistence → display.
func (s *GameService) Play(cfg driving.Config) error {
	processes, err := s.finder.Find(cfg.Patterns)
	if err != nil {
		return fmt.Errorf("process search failed: %w", err)
	}

	if len(processes) == 0 {
		fmt.Printf("No processes found matching %v\n", cfg.Patterns)
		return nil
	}

	fmt.Printf("Found %d process(es) matching %v. Starting game...\n", len(processes), cfg.Patterns)

	board, err := s.store.Load()
	if err != nil {
		board = &score.Board{}
	}

	g := game.New(processes, cfg.ConfirmMode, cfg.Speed, cfg.TimeLimit, s.killer, s.renderer, s.events)
	if err := g.Play(board.HighScore()); err != nil {
		return err
	}

	duration := time.Since(g.StartTime()).Seconds()
	board.Add(score.Entry{
		Kills:    g.Kills(),
		FreedMem: g.FreedMem(),
		Speed:    cfg.Speed,
		Time:     cfg.TimeLimit,
		Duration: duration,
		Date:     time.Now(),
	})
	_ = s.store.Save(board)

	fmt.Printf("\n  Game Over! Kills: %d | Freed: %s | Time: %.1fs\n",
		g.Kills(), util.FormatBytes(g.FreedMem()), duration)

	if g.Kills() > 0 && g.Kills() >= board.HighScore() {
		fmt.Println("  🏆 New high score!")
	}

	board.PrintScores()

	return nil
}
