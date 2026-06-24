package game

import (
	"fmt"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/process"
	"github.com/eirikur-ari/pidshooter/internal/util"
	"github.com/gdamore/tcell/v2"
)

const (
	targetFPS     = 20
	frameDuration = time.Second / targetFPS
)

func (g *Game) init() error {
	if g.screen == nil {
		screen, err := tcell.NewScreen()
		if err != nil {
			return fmt.Errorf("failed to create screen: %w", err)
		}
		g.screen = screen
	}

	if err := g.screen.Init(); err != nil {
		return fmt.Errorf("failed to initialize screen: %w", err)
	}

	g.screen.EnableMouse()
	g.screen.SetStyle(tcell.StyleDefault)
	g.screen.Clear()
	return nil
}

func (g *Game) populateEntities(processes []process.Info) {
	w, h := g.screen.Size()
	for _, p := range processes {
		g.targets = append(g.targets, NewTarget(p, w, h))
	}
}

func (g *Game) run() {
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

func (g *Game) update() {
	w, h := g.screen.Size()

	if g.timeLimit > 0 && g.timeRemaining() == 0 {
		g.running = false
		return
	}

	allDead := true
	for _, e := range g.targets {
		e.Update(w, h, g.speed)
		if e.State != Dead {
			allDead = false
		}
	}

	if allDead && len(g.targets) > 0 {
		g.running = false
	}
}

func (g *Game) render() {
	g.screen.Clear()

	w, h := g.screen.Size()

	aliveStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	killStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)

	for _, e := range g.targets {
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

	memStr := fmt.Sprintf(" FREED: %s ", util.FormatBytes(g.freedMem))
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
		for _, e := range g.targets {
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

func (g *Game) killTarget(e *Target) {
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

func (g *Game) cleanup() {
	if g.screen != nil {
		g.screen.Fini()
		g.screen = nil
	}
}