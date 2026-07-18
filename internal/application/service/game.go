package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/handler"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
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
	processes, err := s.findProcesses(cfg.Patterns)
	if err != nil {
		return err
	}

	if len(processes) == 0 {
		return nil
	}

	board, success := s.loadScoreBoard()

	g := s.newGame(cfg, processes, board)

	endTime, err := s.runLoop(g)
	if err != nil {
		return err
	}

	kills := g.Kills()
	freedMem := g.FreedMem()
	duration := endTime.Sub(g.StartTime()).Seconds()

	s.recordScore(cfg, board, kills, freedMem, duration, success)

	s.printResults(kills, freedMem, duration, board)

	return nil
}

func (s *GameService) findProcesses(patterns []string) ([]process.Info, error) {
	processes, err := s.process.Find(patterns)

	if err != nil {
		return nil, fmt.Errorf("process search failed: %w", err)
	}

	if len(processes) == 0 {
		fmt.Printf("No processes found matching %v\n", patterns)
		return nil, nil
	}

	fmt.Printf("Found %d process(es) matching %v. Starting game...\n", len(processes), patterns)
	return processes, nil
}

func (s *GameService) loadScoreBoard() (*score.Board, bool) {
	board, err := s.store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load scores: %v\n", err)
		return &score.Board{}, false
	}
	return board, true
}

func (s *GameService) recordScore(cfg inbound.GamePlayConfig, board *score.Board, kills int, freedMem int64, duration float64, persist bool) {
	board.Add(score.Entry{
		Kills:    kills,
		FreedMem: freedMem,
		Speed:    cfg.Speed,
		Time:     cfg.TimeLimit,
		Duration: duration,
		Date:     time.Now(),
	})

	if persist {
		if err := s.store.Save(board); err != nil {
			fmt.Fprintf(os.Stderr, "warning: score not saved: %v\n", err)
		}
	}
}

func (s *GameService) printResults(kills int, freedMem int64, duration float64, board *score.Board) {
	fmt.Printf("\n  Game Over! Kills: %d | Freed: %s | Time: %.1fs\n",
		kills, util.FormatBytes(freedMem), duration)
	board.PrintHighScore(kills)
	board.PrintScores()
}

func (s *GameService) newGame(cfg inbound.GamePlayConfig, processes []process.Info, board *score.Board) *game.Game {
	g := game.New(processes, game.Config{Confirm: cfg.ConfirmMode, Speed: cfg.Speed, TimeLimit: cfg.TimeLimit})
	g.SetHighScore(board.HighScore())
	return g
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
	evt := event.NewDispatcher(handler.NewHandler(g))

	for g.IsRunning() {
		s.applyKills(g)
		s.drainEvents(evt, done)
		w, h = s.renderer.Size()
		g.Update(w, h)
		s.renderer.Render(buildFrame(g))
		<-ticker.C
	}
	// Capture end time before deferred Cleanup() runs.
	return time.Now(), nil
}

func (s *GameService) applyKills(g *game.Game) {
	for {
		select {
		case t := <-s.kills:
			g.Kill(t)
		default:
			return
		}
	}
}

func buildFrame(g *game.Game) outbound.FrameState {
	snap := g.Snapshot()
	views := make([]outbound.TargetViewState, len(snap.Targets))
	for i, s := range snap.Targets {
		views[i] = outbound.TargetViewState{X: s.X, Y: s.Y, Tag: s.Tag, Killing: s.Killing}
	}

	return outbound.FrameState{
		Targets: views,
		HUD: outbound.HUDState{
			FreedMem:  g.FreedMem(),
			Kills:     g.Kills(),
			HighScore: g.HighScore(),
		},
		StatusBar: outbound.StatusState{
			Alive:      snap.Alive,
			Speed:      g.Speed(),
			TimeLimit:  g.TimeLimit(),
			TimeLeft:   g.TimeLeft(),
			Confirming: toConfirmViewState(g.ConfirmTarget()),
		},
	}
}

func toConfirmViewState(t *game.Target) *outbound.ConfirmViewState {
	if t == nil {
		return nil
	}
	return &outbound.ConfirmViewState{PID: t.Pid, Name: t.Name}
}

func (s *GameService) drainEvents(d *event.Dispatcher, done <-chan struct{}) {
	for {
		select {
		case ev := <-s.events.Events():
			target := d.Dispatch(ev)
			if target == nil {
				continue
			}
			go func() {
				if killed, err := s.process.Kill(target.Pid, target.Name); err == nil && killed {
					select {
					case s.kills <- target:
					case <-done:
					}
				}
			}()
		default:
			return
		}
	}
}
