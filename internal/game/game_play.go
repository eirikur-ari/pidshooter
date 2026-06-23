// Package game implements the terminal-based PID shooter game loop.
package game

import (
	"fmt"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/process"
	"github.com/gdamore/tcell/v2"
)

const (
	targetFPS     = 20
	frameDuration = time.Second / targetFPS
)

// Game manages the terminal display, entities, and user interaction.
type Game struct {
	screen      tcell.Screen
	entities    []*Entity
	confirmMode bool
	speed       float64
	timeLimit   int
	running     bool
	confirming  *Entity
	Session
}

// New creates a new game instance.
func New(processes []process.Info, confirmMode bool, speed float64, timeLimit int) *Game {
	return &Game{
		entities:    make([]*Entity, 0, len(processes)),
		confirmMode: confirmMode,
		speed:       speed,
		running:     true,
		timeLimit:   timeLimit,
	}
}

// Stop signals the game to exit.
func (g *Game) Stop() { g.running = false }

// Init initializes the terminal screen.
func (g *Game) Init() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("failed to create screen: %w", err)
	}

	if err := screen.Init(); err != nil {
		return fmt.Errorf("failed to initialize screen: %w", err)
	}

	screen.EnableMouse()
	screen.SetStyle(tcell.StyleDefault)
	screen.Clear()

	g.screen = screen
	return nil
}

// PopulateEntities creates entity objects for each process.
func (g *Game) PopulateEntities(processes []process.Info) {
	w, h := g.screen.Size()
	for _, p := range processes {
		g.entities = append(g.entities, NewEntity(p, w, h))
	}
}

// Run executes the main game loop.
func (g *Game) Run() {
	g.startTime = time.Now()
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	eventCh := make(chan tcell.Event, 10)
	go func() {
		for {
			ev := g.screen.PollEvent()
			if ev == nil {
				return
			}
			eventCh <- ev
		}
	}()

	for g.running {
		g.drainEvents(eventCh)
		g.update()
		g.render()
		<-ticker.C
	}
}

func (g *Game) drainEvents(eventCh <-chan tcell.Event) {
	for {
		select {
		case ev := <-eventCh:
			g.handleEvent(ev)
		default:
			return
		}
	}
}

func (g *Game) handleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventMouse:
		if ev.Buttons() == tcell.Button1 {
			x, y := ev.Position()
			g.handleMouseClick(x, y)
		}
	case *tcell.EventKey:
		g.HandleKeyPress(ev.Key(), ev.Rune())
	case *tcell.EventResize:
		g.screen.Sync()
	}
}

func (g *Game) handleMouseClick(x, y int) {
	if g.confirming != nil {
		return
	}

	for _, e := range g.entities {
		if e.Contains(x, y) {
			if g.confirmMode {
				g.confirming = e
			} else {
				g.killEntity(e)
			}
			return
		}
	}
}

// HandleKeyPress processes keyboard input.
func (g *Game) HandleKeyPress(key tcell.Key, r rune) {
	switch {
	case key == tcell.KeyEscape:
		g.running = false
	case key == tcell.KeyCtrlC:
		g.running = false
	case key == tcell.KeyCtrlZ:
		g.running = false
	case r == 'q' || r == 'Q':
		if g.confirming != nil {
			g.confirming = nil
		} else {
			g.running = false
		}
	case r == 'y' || r == 'Y':
		if g.confirming != nil {
			g.killEntity(g.confirming)
			g.confirming = nil
		}
	case r == 'n' || r == 'N':
		if g.confirming != nil {
			g.confirming = nil
		}
	case r == '+':
		g.speed += 0.5
		if g.speed > 5.0 {
			g.speed = 5.0
		}
	case r == '-':
		g.speed -= 0.5
		if g.speed < 0.1 {
			g.speed = 0.1
		}
	}
}

func (g *Game) update() {
	w, h := g.screen.Size()

	if g.timeLimit > 0 && g.timeRemaining() == 0 {
		g.running = false
		return
	}

	allDead := true
	for _, e := range g.entities {
		e.Update(w, h, g.speed)
		if e.State != Dead {
			allDead = false
		}
	}

	if allDead && len(g.entities) > 0 {
		g.running = false
	}
}

func (g *Game) render() {
	g.screen.Clear()

	w, h := g.screen.Size()

	aliveStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	killStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)

	for _, e := range g.entities {
		if e.State == Dead {
			continue
		}

		label := e.Label()
		x := int(e.PosX)
		y := int(e.PosY)

		style := aliveStyle
		if e.State == Killing {
			style = killStyle
		}

		for i, ch := range label {
			if x+i < w && y < h-1 {
				g.screen.SetContent(x+i, y, ch, nil, style)
			}
		}
	}

	memStr := fmt.Sprintf(" FREED: %s ", FormatBytes(g.freedMem))
	memStyle := tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true)
	for i, ch := range memStr {
		if i < w {
			g.screen.SetContent(i, 0, ch, nil, memStyle)
		}
	}

	hiStr := fmt.Sprintf(" Highscore: %d ", g.highScore)
	hiStyle := tcell.StyleDefault.Foreground(tcell.ColorPurple).Bold(true)
	hiX := (w - len(hiStr)) / 2
	if hiX < 0 {
		hiX = 0
	}
	for i, ch := range hiStr {
		if hiX+i < w {
			g.screen.SetContent(hiX+i, 0, ch, nil, hiStyle)
		}
	}

	scoreStr := fmt.Sprintf(" KILLS: %d ", g.kills)
	scoreStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	scoreX := w - len(scoreStr)
	if scoreX < 0 {
		scoreX = 0
	}
	for i, ch := range scoreStr {
		if scoreX+i < w {
			g.screen.SetContent(scoreX+i, 0, ch, nil, scoreStyle)
		}
	}

	g.drawStatusBar(w, h)
	g.screen.Show()
}

func (g *Game) drawStatusBar(w, h int) {
	statusStyle := tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite)

	for x := 0; x < w; x++ {
		g.screen.SetContent(x, h-1, ' ', nil, statusStyle)
	}

	var status string
	if g.confirming != nil {
		status = fmt.Sprintf(" Kill [%d %s]? (Y)es / (N)o / (Q)uit",
			g.confirming.Pid(), g.confirming.Name())
	} else {
		alive := 0
		for _, e := range g.entities {
			if e.State == Alive {
				alive++
			}
		}
		timerStr := ""
		if g.timeLimit > 0 {
			timerStr = fmt.Sprintf(" | Time: %ds", int(g.timeRemaining().Seconds()))
		}
		status = fmt.Sprintf(" Targets: %d | Speed: %.1fx%s | Click to kill | +/- speed | 'q' quit", alive, g.speed, timerStr)
	}

	for i, ch := range status {
		if i < w {
			g.screen.SetContent(i, h-1, ch, nil, statusStyle)
		}
	}
}

func (g *Game) killEntity(e *Entity) {
	if e.State != Alive {
		return
	}
	_ = e.Kill()
	e.StartKillAnim()
	g.Session.RecordKill(e.Rss())
}

func (g *Game) timeRemaining() time.Duration {
	remaining := time.Duration(g.timeLimit)*time.Second - time.Since(g.startTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Cleanup restores the terminal. Safe to call multiple times.
func (g *Game) Cleanup() {
	if g.screen != nil {
		g.screen.Fini()
		g.screen = nil
	}
}
