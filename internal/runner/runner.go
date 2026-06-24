// Package runner orchestrates a full game session: process discovery, game lifecycle, and score persistence.
package runner

import (
	"fmt"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/game"
	"github.com/eirikur-ari/pidshooter/internal/process"
	"github.com/eirikur-ari/pidshooter/internal/score"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// Config holds the parameters needed to start a game session.
type Config struct {
	Patterns    []string
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}

// Start runs a complete game session from process discovery through score persistence.
func Start(cfg Config) error {
	return start(cfg, process.NewFinder())
}

func start(cfg Config, finder process.Finder) error {
	processes, err := finder.Find(cfg.Patterns)
	if err != nil {
		return fmt.Errorf("process search failed: %w", err)
	}

	if len(processes) == 0 {
		fmt.Printf("No processes found matching %v\n", cfg.Patterns)
		return nil
	}

	fmt.Printf("Found %d process(es) matching %v. Starting game...\n", len(processes), cfg.Patterns)

	scoreBoard := score.Load()
	g := game.New(processes, cfg.ConfirmMode, cfg.Speed, cfg.TimeLimit)
	if err := g.Play(scoreBoard.HighScore()); err != nil {
		return err
	}

	duration := time.Since(g.StartTime()).Seconds()
	scoreBoard.Add(score.Entry{
		Kills:    g.Kills(),
		FreedMem: g.FreedMem(),
		Speed:    cfg.Speed,
		Time:     cfg.TimeLimit,
		Duration: duration,
		Date:     time.Now(),
	})
	_ = scoreBoard.Save()

	fmt.Printf("\n  Game Over! Kills: %d | Freed: %s | Time: %.1fs\n",
		g.Kills(), util.FormatBytes(g.FreedMem()), duration)
	if g.Kills() > 0 && g.Kills() >= scoreBoard.HighScore() {
		fmt.Println("  🏆 New high score!")
	}
	scoreBoard.PrintScores()

	return nil
}