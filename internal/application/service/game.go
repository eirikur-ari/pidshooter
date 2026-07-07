package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

const frameDuration = time.Second / 20

// GameService implements contract.GamePlay by orchestrating core domain objects and outbound ports.
type GameService struct {
	finder   contract.Finder
	killer   contract.ProcessKiller
	store    contract.Store
	renderer contract.Renderer
	events   contract.EventSource
}

// NewGameService constructs a GameService with all required outbound ports injected.
func NewGameService(
	finder contract.Finder,
	killer contract.ProcessKiller,
	store contract.Store,
	renderer contract.Renderer,
	events contract.EventSource,
) *GameService {
	return &GameService{
		finder:   finder,
		killer:   killer,
		store:    store,
		renderer: renderer,
		events:   events,
	}
}

// Play runs a complete game session: discovery → game loop → score persistence → display.
func (s *GameService) Play(cfg contract.GamePlayConfig) error {
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

	g := game.New(processes, cfg.ConfirmMode, cfg.Speed, cfg.TimeLimit)
	g.SetHighScore(board.HighScore())

	if err := s.runLoop(g); err != nil {
		return err
	}

	kills := g.Kills()
	freedMem := g.FreedMem()
	duration := time.Since(g.StartTime()).Seconds()

	board.Add(score.Entry{
		Kills:    kills,
		FreedMem: freedMem,
		Speed:    cfg.Speed,
		Time:     cfg.TimeLimit,
		Duration: duration,
		Date:     time.Now(),
	})
	if err := s.store.Save(board); err != nil {
		fmt.Fprintf(os.Stderr, "warning: score not saved: %v\n", err)
	}

	fmt.Printf("\n  Game Over! Kills: %d | Freed: %s | Time: %.1fs\n",
		kills, util.FormatBytes(freedMem), duration)
	board.PrintHighScore(kills)
	board.PrintScores()

	return nil
}

func (s *GameService) runLoop(g *game.Game) error {
	if err := s.renderer.Init(); err != nil {
		return fmt.Errorf("renderer initialization failed: %w", err)
	}
	defer s.renderer.Cleanup()

	w, h := s.renderer.Size()
	g.Init(w, h)

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

	for g.Running() {
		s.drainEvents(g)
		w, h = s.renderer.Size()
		g.Update(w, h)
		s.renderer.Render(g.Frame())
		<-ticker.C
	}
	return nil
}

func (s *GameService) drainEvents(g *game.Game) {
	for {
		select {
		case ev := <-s.events.Events():
			var req *game.KillRequest
			switch ev := ev.(type) {
			case game.ClickEvent:
				req = g.HandleClick(ev.X, ev.Y)
			case game.KeyEvent:
				switch ev.Key {
				case game.KeyEscape, game.KeyCtrlC, game.KeyCtrlZ:
					g.Stop()
				default:
					req = g.HandleKey(ev.Ch)
				}
			}
			if req != nil {
				if s.killer.Kill(req.Target.Pid, req.Target.Name) == nil {
					g.CompleteKill(req.Target)
				}
			}
		default:
			return
		}
	}
}
