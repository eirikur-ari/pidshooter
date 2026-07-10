package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

const frameDuration = time.Second / 20

// GameService implements inbound.GamePlay by orchestrating core domain objects and outbound ports.
type GameService struct {
	process  outbound.Process
	store    outbound.ScoreStore
	renderer outbound.Renderer
	events   outbound.InputSource
	kills    chan *game.Target
}

// NewGameService constructs a GameService with all required outbound ports injected.
func NewGameService(
	process outbound.Process,
	store outbound.ScoreStore,
	renderer outbound.Renderer,
	events outbound.InputSource,
) *GameService {
	return &GameService{
		process:  process,
		store:    store,
		renderer: renderer,
		events:   events,
	}
}

// Play runs a complete game session: discovery → game loop → score persistence → display.
func (s *GameService) Play(cfg inbound.GamePlayConfig) error {
	processes, err := s.process.Find(cfg.Patterns)
	if err != nil {
		return fmt.Errorf("process search failed: %w", err)
	}

	if len(processes) == 0 {
		fmt.Printf("No processes found matching %v\n", cfg.Patterns)
		return nil
	}

	fmt.Printf("Found %d process(es) matching %v. Starting game...\n", len(processes), cfg.Patterns)

	board, loadErr := s.store.Load()
	if loadErr != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load scores: %v\n", loadErr)
		board = &score.Board{}
	}

	g := game.New(processes, game.Config{ConfirmMode: cfg.ConfirmMode, Speed: cfg.Speed, TimeLimit: cfg.TimeLimit})
	g.SetHighScore(board.HighScore())

	endTime, err := s.runLoop(g)
	if err != nil {
		return err
	}

	kills := g.Kills()
	freedMem := g.FreedMem()
	duration := endTime.Sub(g.StartTime()).Seconds()

	board.Add(score.Entry{
		Kills:    kills,
		FreedMem: freedMem,
		Speed:    cfg.Speed,
		Time:     cfg.TimeLimit,
		Duration: duration,
		Date:     time.Now(),
	})
	if loadErr == nil {
		if err := s.store.Save(board); err != nil {
			fmt.Fprintf(os.Stderr, "warning: score not saved: %v\n", err)
		}
	}

	fmt.Printf("\n  Game Over! Kills: %d | Freed: %s | Time: %.1fs\n",
		kills, util.FormatBytes(freedMem), duration)
	board.PrintHighScore(kills)
	board.PrintScores()

	return nil
}

func (s *GameService) runLoop(g *game.Game) (time.Time, error) {
	if err := s.renderer.Init(); err != nil {
		return time.Time{}, fmt.Errorf("renderer initialization failed: %w", err)
	}
	defer s.renderer.Cleanup()

	w, h := s.renderer.Size()
	g.Start(w, h)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	defer signal.Stop(sigCh)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-sigCh:
			g.Stop()
		case <-done:
		}
	}()

	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	s.kills = make(chan *game.Target, 10)

	for g.IsRunning() {
		s.applyKills(g)
		s.drainEvents(g, done)
		w, h = s.renderer.Size()
		g.Update(w, h)
		s.renderer.Render(g.Frame())
		<-ticker.C
	}
	// Capture end time before deferred Cleanup() runs.
	return time.Now(), nil
}

func (s *GameService) applyKills(g *game.Game) {
	for {
		select {
		case t := <-s.kills:
			g.CompleteKill(t)
		default:
			return
		}
	}
}

func (s *GameService) drainEvents(g *game.Game, done <-chan struct{}) {
	for {
		select {
		case ev := <-s.events.Events():
			var target *game.Target
			switch ev := ev.(type) {
			case event.ClickEvent:
				target = g.HandleClick(ev.X, ev.Y)
			case event.KeyEvent:
				switch ev.Key {
				case event.KeyEscape, event.KeyCtrlC, event.KeyCtrlZ:
					g.Stop()
				default:
					target = g.HandleKey(ev.Ch)
				}
			}
			if target != nil {
				go func() {
					if killed, err := s.process.Kill(target.Pid, target.Name); err == nil && killed {
						select {
						case s.kills <- target:
						case <-done:
						}
					}
				}()
			}
		default:
			return
		}
	}
}
