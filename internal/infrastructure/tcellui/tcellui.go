// Package tcellui implements the outbound.Renderer and outbound.InputSource ports using tcell.
package tcellui

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// UI implements both outbound.Renderer and outbound.InputSource using a tcell.Screen.
// The event poll goroutine is started inside Init() after the screen is ready.
type UI struct {
	screen      tcell.Screen
	ch          chan outbound.InputEvent
	done        chan struct{}
	cleanupOnce sync.Once
}

// NewUI returns a tcellui.UI wrapping the given screen.
// The caller must not call tcell.Screen.Init directly; use UI.Init() instead.
func NewUI(screen tcell.Screen) *UI {
	return &UI{
		screen: screen,
		ch:     make(chan outbound.InputEvent, 10),
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
// Safe to call more than once; only the first call has any effect.
func (a *UI) Cleanup() {
	a.cleanupOnce.Do(func() {
		close(a.done)
		a.screen.Fini()
	})
}

// Size returns the current terminal dimensions.
func (a *UI) Size() (int, int) {
	return a.screen.Size()
}

// Render translates an outbound.FrameState into tcell draw calls.
func (a *UI) Render(state outbound.FrameState) {
	a.screen.Clear()
	w, h := a.screen.Size()

	aliveStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	killStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)

	for _, tv := range state.Targets {
		style := aliveStyle
		if tv.Killing {
			style = killStyle
		}
		col := 0
		for _, ch := range tv.Tag {
			if tv.X+col < w && tv.Y < h-1 {
				a.screen.SetContent(tv.X+col, tv.Y, ch, nil, style)
			}
			col++
		}
	}

	a.drawHUD(w, state.HUD)
	a.drawStatusBar(w, h, state.StatusBar)
	a.screen.Show()
}

// Events returns the channel of translated game input events.
func (a *UI) Events() <-chan outbound.InputEvent {
	return a.ch
}

func (a *UI) drawHUD(w int, hud outbound.HUDState) {
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

	scoreStr := fmt.Sprintf(" KILLS: %d ", hud.Kills)
	scoreStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	scoreX := w - len(scoreStr)
	if scoreX < 0 {
		scoreX = 0
	}

	if hiX >= len(memStr) && hiX+len(hiStr) <= scoreX {
		for i, ch := range hiStr {
			if hiX+i < w {
				a.screen.SetContent(hiX+i, 0, ch, nil, hiStyle)
			}
		}
	}

	for i, ch := range scoreStr {
		if scoreX+i < w {
			a.screen.SetContent(scoreX+i, 0, ch, nil, scoreStyle)
		}
	}
}

func (a *UI) drawStatusBar(w, h int, status outbound.StatusState) {
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

func (a *UI) poll() {
	for {
		ev := a.screen.PollEvent()
		if ev == nil {
			return
		}
		var inputEvent outbound.InputEvent
		switch ev := ev.(type) {
		case *tcell.EventMouse:
			if ev.Buttons() != tcell.Button1 {
				continue
			}
			x, y := ev.Position()
			inputEvent = outbound.ClickEvent{X: x, Y: y}
		case *tcell.EventKey:
			ie, ok := translateEvent(ev)
			if !ok {
				continue
			}
			inputEvent = ie
		case *tcell.EventResize:
			a.screen.Sync()
			continue
		default:
			continue
		}
		select {
		case a.ch <- inputEvent:
		case <-a.done:
			return
		}
	}
}

func translateEvent(ev *tcell.EventKey) (outbound.InputEvent, bool) {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		return outbound.QuitEvent{}, true
	}
	switch ev.Rune() {
	case 'q', 'Q':
		return outbound.QuitEvent{}, true
	case 'y', 'Y':
		return outbound.ConfirmEvent{Accept: true}, true
	case 'n', 'N':
		return outbound.ConfirmEvent{Accept: false}, true
	case '+', '=':
		return outbound.SpeedEvent{Faster: true}, true
	case '-', '_':
		return outbound.SpeedEvent{Faster: false}, true
	}
	return nil, false
}
