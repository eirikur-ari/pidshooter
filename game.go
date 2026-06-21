package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	// targetFPS is the desired frames per second for the animation.
	targetFPS = 20
	// frameDuration is the time between frames.
	frameDuration = time.Second / targetFPS
)

// Game manages the terminal display, entities, and user interaction.
type Game struct {
	screen      tcell.Screen
	entities    []*Entity
	confirmMode bool
	speed       float64
	running     bool
	confirming  *Entity // Entity pending kill confirmation
	kills       int     // Kill score counter
	freedMem    int64   // Total freed memory in bytes
	timeLimit   int     // Time limit in seconds (0 = no limit)
	startTime   time.Time
	highScore   int     // Top high score to display
}

// NewGame creates a new game instance with the given processes.
func NewGame(processes []ProcessInfo, confirmMode bool, speed float64, timeLimit int) *Game {
	return &Game{
		entities:    make([]*Entity, 0, len(processes)),
		confirmMode: confirmMode,
		speed:       speed,
		running:     true,
		timeLimit:   timeLimit,
	}
}

// Init initializes the terminal screen and creates entities.
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
// Must be called after Init() so screen dimensions are available.
func (g *Game) PopulateEntities(processes []ProcessInfo) {
	w, h := g.screen.Size()
	for _, p := range processes {
		g.entities = append(g.entities, NewEntity(p.PID, p.Name, p.RSS, w, h))
	}
}

// Run executes the main game loop.
func (g *Game) Run() {
	g.startTime = time.Now()
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	// Event channel for non-blocking event polling
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
		// Process all pending events
		g.drainEvents(eventCh)

		// Update game state
		g.update()

		// Render frame
		g.render()

		// Wait for next frame
		<-ticker.C
	}
}

// drainEvents processes all queued events without blocking.
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

// handleEvent dispatches events to appropriate handlers.
func (g *Game) handleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventMouse:
		// Only handle on button press, ignore release (Buttons() == 0) and drag
		if ev.Buttons() == tcell.Button1 {
			x, y := ev.Position()
			g.handleMouseClick(x, y)
		}
	case *tcell.EventKey:
		g.handleKeyPress(ev.Key(), ev.Rune())
	case *tcell.EventResize:
		g.screen.Sync()
	}
}

// handleMouseClick checks if any entity was clicked and initiates kill.
func (g *Game) handleMouseClick(x, y int) {
	// If awaiting confirmation, only Y/N keys work — ignore mouse clicks
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

// handleConfirmClick is no longer needed — confirmation handled via key or second click.

// handleKeyPress processes keyboard input.
func (g *Game) handleKeyPress(key tcell.Key, r rune) {
	switch {
	case key == tcell.KeyEscape:
		g.running = false
	case key == tcell.KeyCtrlC:
		g.running = false
	case key == tcell.KeyCtrlZ:
		g.running = false
	case r == 'q' || r == 'Q':
		if g.confirming != nil {
			// Cancel confirmation
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
	case r == '+' || r == '=':
		g.speed += 0.5
		if g.speed > 5.0 {
			g.speed = 5.0
		}
	case r == '-' || r == '_':
		g.speed -= 0.5
		if g.speed < 0.1 {
			g.speed = 0.1
		}
	}
}

// update advances the game state by one frame.
func (g *Game) update() {
	w, h := g.screen.Size()

	// Check timer
	if g.timeLimit > 0 {
		elapsed := time.Since(g.startTime)
		if elapsed >= time.Duration(g.timeLimit)*time.Second {
			g.running = false
			return
		}
	}

	allDead := true
	for _, e := range g.entities {
		e.Update(w, h, g.speed)
		if e.State != StateDead {
			allDead = false
		}
	}

	// Exit when all processes are eliminated
	if allDead && len(g.entities) > 0 {
		g.running = false
	}
}

// render draws all entities and the status bar.
func (g *Game) render() {
	g.screen.Clear()

	w, h := g.screen.Size()

	// Draw entities
	aliveStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	killStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)

	for _, e := range g.entities {
		if e.State == StateDead {
			continue
		}

		label := e.Label()
		x := int(e.X)
		y := int(e.Y)

		style := aliveStyle
		if e.State == StateKilling {
			style = killStyle
		}

		for i, ch := range label {
			if x+i < w && y < h-1 {
				g.screen.SetContent(x+i, y, ch, nil, style)
			}
		}
	}

	// Draw freed memory in top-left corner
	memStr := fmt.Sprintf(" FREED: %s ", formatBytes(g.freedMem))
	memStyle := tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true)
	for i, ch := range memStr {
		if i < w {
			g.screen.SetContent(i, 0, ch, nil, memStyle)
		}
	}

	// Draw high score in top-center
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

	// Draw score in top-right corner
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

	// Draw status bar
	g.drawStatusBar(w, h)

	g.screen.Show()
}

// drawStatusBar renders the bottom status line.
func (g *Game) drawStatusBar(w, h int) {
	statusStyle := tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite)

	// Clear status line
	for x := 0; x < w; x++ {
		g.screen.SetContent(x, h-1, ' ', nil, statusStyle)
	}

	var status string
	if g.confirming != nil {
		status = fmt.Sprintf(" Kill [%d %s]? (Y)es / (N)o / (Q)uit",
			g.confirming.PID, g.confirming.Name)
	} else {
		alive := 0
		for _, e := range g.entities {
			if e.State == StateAlive {
				alive++
			}
		}
		timerStr := ""
		if g.timeLimit > 0 {
			remaining := time.Duration(g.timeLimit)*time.Second - time.Since(g.startTime)
			if remaining < 0 {
				remaining = 0
			}
			timerStr = fmt.Sprintf(" | Time: %ds", int(remaining.Seconds()))
		}
		status = fmt.Sprintf(" Targets: %d | Speed: %.1fx%s | Click to kill | +/- speed | 'q' quit", alive, g.speed, timerStr)
	}

	for i, ch := range status {
		if i < w {
			g.screen.SetContent(i, h-1, ch, nil, statusStyle)
		}
	}
}

// killEntity sends SIGKILL to the process and starts the kill animation.
func (g *Game) killEntity(e *Entity) {
	if e.State != StateAlive {
		return
	}

	// Attempt to kill the process
	proc, err := os.FindProcess(e.PID)
	if err == nil {
		// Send SIGKILL - ignore errors (process may have already exited)
		_ = proc.Signal(syscall.SIGKILL)
	}

	e.StartKillAnim()
	g.kills++
	g.freedMem += e.RSS
}

// Cleanup restores the terminal to its original state. Safe to call multiple times.
func (g *Game) Cleanup() {
	if g.screen != nil {
		g.screen.Fini()
		g.screen = nil
	}
}

// formatBytes formats a byte count into a human-readable string.
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
