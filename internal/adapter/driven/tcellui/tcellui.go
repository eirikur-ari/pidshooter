// Package tcellui implements the game Renderer and EventSource ports using tcell.
package tcellui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// UI implements both driven.Renderer and driven.EventSource using a tcell.Screen.
// The event poll goroutine is started inside Init() after the screen is ready.
type UI struct {
	screen tcell.Screen
	ch     chan driven.InputEvent
	done   chan struct{}
}

// New returns a tcellui.UI wrapping the given screen.
// The caller must not call tcell.Screen.Init directly; use UI.Init() instead.
func New(screen tcell.Screen) *UI {
	return &UI{
		screen: screen,
		ch:     make(chan driven.InputEvent, 10),
		done:   make(chan struct{}),
	}
}

// Init initialises the screen and starts the event polling goroutine.
func (a *UI) Init() error {
	if err := a.screen.Init(); err != nil {
		return fmt.Errorf("failed to initialize screen: %w", err)
	}
	a.screen.EnableMouse()
	a.screen.SetStyle(tcell.StyleDefault)
	a.screen.Clear()
	go a.poll()
	return nil
}

// Cleanup signals the poll goroutine to stop, then shuts down the screen.
func (a *UI) Cleanup() {
	close(a.done)
	a.screen.Fini()
}

// Size returns the current terminal dimensions.
func (a *UI) Size() (int, int) {
	return a.screen.Size()
}

// Render translates a domain Frame into tcell draw calls.
func (a *UI) Render(frame driven.Frame) {
	a.screen.Clear()
	w, h := a.screen.Size()

	aliveStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	killStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)

	for _, tv := range frame.Targets {
		style := aliveStyle
		if tv.Killing {
			style = killStyle
		}
		col := 0
		for _, ch := range tv.Label {
			if tv.X+col < w && tv.Y < h-1 {
				a.screen.SetContent(tv.X+col, tv.Y, ch, nil, style)
			}
			col++
		}
	}

	a.drawHUD(w, frame.HUD)
	a.drawStatusBar(w, h, frame.StatusBar)
	a.screen.Show()
}

func (a *UI) drawHUD(w int, hud driven.HUDState) {
	memStr := fmt.Sprintf(" FREED: %s ", util.FormatBytes(hud.FreedMem))
	memStyle := tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true)
	for i, ch := range memStr {
		if i < w {
			a.screen.SetContent(i, 0, ch, nil, memStyle)
		}
	}

	hiStr := fmt.Sprintf(" Highscore: %d ", hud.HighScore)
	hiStyle := tcell.StyleDefault.Foreground(tcell.ColorPurple).Bold(true)
	hiX := (w - len(hiStr)) / 2
	if hiX < 0 {
		hiX = 0
	}
	for i, ch := range hiStr {
		if hiX+i < w {
			a.screen.SetContent(hiX+i, 0, ch, nil, hiStyle)
		}
	}

	scoreStr := fmt.Sprintf(" KILLS: %d ", hud.Kills)
	scoreStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	scoreX := w - len(scoreStr)
	if scoreX < 0 {
		scoreX = 0
	}
	for i, ch := range scoreStr {
		if scoreX+i < w {
			a.screen.SetContent(scoreX+i, 0, ch, nil, scoreStyle)
		}
	}
}

func (a *UI) drawStatusBar(w, h int, status driven.StatusState) {
	statusStyle := tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite)

	for x := 0; x < w; x++ {
		a.screen.SetContent(x, h-1, ' ', nil, statusStyle)
	}

	var statusStr string
	if status.Confirming != nil {
		statusStr = fmt.Sprintf(" Kill [%d %s]? (Y)es / (N)o / (Q)uit",
			status.Confirming.PID, status.Confirming.Name)
	} else {
		timerStr := ""
		if status.TimeLimit > 0 {
			timerStr = fmt.Sprintf(" | Time: %ds", status.TimeLeft)
		}
		statusStr = fmt.Sprintf(" Targets: %d | Speed: %.1fx%s | Click to kill | +/- speed | 'q' quit",
			status.Alive, status.Speed, timerStr)
	}

	for i, ch := range statusStr {
		if i < w {
			a.screen.SetContent(i, h-1, ch, nil, statusStyle)
		}
	}
}

// Events returns the channel of translated domain input events.
func (a *UI) Events() <-chan driven.InputEvent {
	return a.ch
}

func (a *UI) poll() {
	for {
		ev := a.screen.PollEvent()
		if ev == nil {
			return
		}
		var event driven.InputEvent
		switch ev := ev.(type) {
		case *tcell.EventMouse:
			if ev.Buttons() != tcell.Button1 {
				continue
			}
			x, y := ev.Position()
			event = driven.ClickEvent{X: x, Y: y}
		case *tcell.EventKey:
			event = driven.KeyEvent{Key: translateKey(ev.Key()), Ch: ev.Rune()}
		case *tcell.EventResize:
			a.screen.Sync()
			event = driven.ResizeEvent{}
		default:
			continue
		}
		select {
		case a.ch <- event:
		case <-a.done:
			return
		}
	}
}

func translateKey(k tcell.Key) driven.KeyCode {
	switch k {
	case tcell.KeyEscape:
		return driven.KeyEscape
	case tcell.KeyCtrlC:
		return driven.KeyCtrlC
	case tcell.KeyCtrlZ:
		return driven.KeyCtrlZ
	default:
		return driven.KeyNone
	}
}
