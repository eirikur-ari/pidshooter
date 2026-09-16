// Package tcellui implements the outbound.Renderer and outbound.InputSource ports using tcell.
package tcellui

import (
	"errors"
	"fmt"
	"sync"
	"unicode/utf8"

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
	initialized bool
}

// Compile-time assertions that *UI satisfies both ports, so a signature
// drift fails the build here instead of at a distant call site.
var (
	_ outbound.Renderer    = (*UI)(nil)
	_ outbound.InputSource = (*UI)(nil)
)

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
	if a.initialized {
		return errors.New("tcellui: already initialized")
	}
	if err := a.screen.Init(); err != nil {
		return fmt.Errorf("failed to initialize screen: %w", err)
	}
	a.screen.EnableMouse(tcell.MouseButtonEvents)
	a.screen.SetStyle(tcell.StyleDefault)
	a.screen.Clear()
	a.initialized = true
	go a.poll()
	return nil
}

// Cleanup signals the poll goroutine to stop, then shuts down the screen.
// Safe to call more than once; only the first call has any effect. Safe to
// call even if Init was never called or failed.
func (a *UI) Cleanup() {
	a.cleanupOnce.Do(func() {
		close(a.done)
		if a.initialized {
			a.screen.Fini()
		}
	})
}

// WindowSize returns the current terminal dimensions.
func (a *UI) WindowSize() outbound.WindowSize {
	w, h := a.screen.Size()
	return outbound.WindowSize{Width: w, Height: h}
}

// ChromeSize reports that tcellui reserves one row at the top for the
// HUD and one row at the bottom for the status bar.
func (a *UI) ChromeSize() outbound.ChromeSize {
	return outbound.ChromeSize{Top: 1, Bottom: 1}
}

// Render translates an outbound.FrameState into tcell draw calls.
func (a *UI) Render(state outbound.FrameState) {
	a.screen.Clear()
	width, height := a.screen.Size()
	chrome := a.ChromeSize()
	top, bottom := chrome.Top, chrome.Bottom

	aliveStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	killStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	fleeStyle := tcell.StyleDefault.Foreground(tcell.ColorOrange).Bold(true)

	for _, target := range state.Targets {
		style := aliveStyle
		tag := target.Tag
		switch {
		case target.Killing:
			style = killStyle
			tag = killAnimationFrame(target.AnimationProgress)
		case target.Fleeing:
			style = fleeStyle
			tag = fleeAnimationFrame(target.AnimationProgress)
		}
		if target.Y >= top && target.Y < height-bottom {
			a.screen.PutStrStyled(target.X, target.Y, tag, style)
		}
	}

	a.drawHUD(width, state.HUD)
	a.drawStatusBar(width, height, state.StatusBar)
	a.screen.Show()
}

// Events returns the channel of translated game input events.
func (a *UI) Events() <-chan outbound.InputEvent {
	return a.ch
}

func (a *UI) drawHUD(w int, hud outbound.HUDState) {
	memStr := fmt.Sprintf(" FREED: %s ", util.FormatBytes(hud.FreedMem))
	memStyle := tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true)
	memWidth := utf8.RuneCountInString(memStr)

	hiStr := fmt.Sprintf(" Highscore: %d ", hud.HighScore)
	hiStyle := tcell.StyleDefault.Foreground(tcell.ColorPurple).Bold(true)
	hiWidth := utf8.RuneCountInString(hiStr)
	hiX := (w - hiWidth) / 2

	scoreStr := fmt.Sprintf(" KILLS: %d ", hud.Kills)
	scoreStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	scoreWidth := utf8.RuneCountInString(scoreStr)
	scoreX := max(w-scoreWidth, 0)

	a.screen.PutStrStyled(0, 0, truncateRunes(memStr, scoreX), memStyle)

	if hiX >= memWidth && hiX+hiWidth <= scoreX {
		a.screen.PutStrStyled(hiX, 0, hiStr, hiStyle)
	}

	a.screen.PutStrStyled(scoreX, 0, scoreStr, scoreStyle)
}

func (a *UI) drawStatusBar(w, h int, status outbound.StatusState) {
	statusStyle := tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite)

	for x := range w {
		a.screen.SetContent(x, h-1, ' ', nil, statusStyle)
	}

	var statusStr string
	if status.Confirming != nil {
		prefix := fmt.Sprintf(" Kill [%d ", status.Confirming.PID)
		const suffix = "]? (Y)es / (N)o / (Q)uit"
		nameBudget := w - utf8.RuneCountInString(prefix) - utf8.RuneCountInString(suffix)
		statusStr = prefix + truncateWithEllipsis(status.Confirming.Name, nameBudget) + suffix
	} else {
		timerStr := ""
		if status.TimeLimit > 0 {
			timerStr = fmt.Sprintf(" | Time: %ds", status.TimeLeft)
		}
		statusStr = fmt.Sprintf(" Targets: %d | Speed: %.1fx%s | Click to kill | +/- speed | 'q' quit",
			status.Alive, status.Speed, timerStr)
	}

	a.screen.PutStrStyled(0, h-1, statusStr, statusStyle)
}

// truncateRunes returns s truncated to at most n runes, so a caller can give
// PutStrStyled an explicit right-edge budget instead of relying on a
// later draw call to overwrite whatever runs past it.
func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

// truncateWithEllipsis is truncateRunes for text a player reads directly,
// signaling with a trailing "…" that something was cut rather than
// silently dropping it.
func truncateWithEllipsis(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n == 1 {
		return "…"
	}
	return string(runes[:n-1]) + "…"
}

func killAnimationFrame(progress float64) string {
	return animationFrame([]string{"💥", "✦ KILLED ✦", "· · ·", "  ·  ", "     "}, progress)
}

func fleeAnimationFrame(progress float64) string {
	return animationFrame([]string{"🏃💨", "↝ RAN AWAY ↝", "· · ·", "  ·  ", "     "}, progress)
}

// animationFrame picks the frame from frames corresponding to progress, a
// fraction from 0 to 1 through the animation.
func animationFrame(frames []string, progress float64) string {
	idx := int(progress * float64(len(frames)))
	if idx >= len(frames) {
		idx = len(frames) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return frames[idx]
}

func (a *UI) poll() {
	defer close(a.ch)
	for {
		pollEvent := a.screen.PollEvent()
		if pollEvent == nil {
			return
		}
		var inputEvent outbound.InputEvent
		switch event := pollEvent.(type) {
		case *tcell.EventMouse:
			// Strict equality: a chord (e.g. Button1+Button2 held together)
			// is deliberately dropped rather than treated as a click.
			if event.Buttons() != tcell.Button1 {
				continue
			}
			x, y := event.Position()
			_, height := a.screen.Size()
			chrome := a.ChromeSize()
			if y < chrome.Top || y >= height-chrome.Bottom {
				continue // window chrome rows never contain a target
			}
			inputEvent = outbound.ClickEvent{X: x, Y: y}
		case *tcell.EventKey:
			ie, ok := translateEvent(event)
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

// translateEvent maps a recognized key press to a game input event. A key
// matches by its rune or Key value alone; modifiers (e.g. Alt, Shift) are
// not considered.
func translateEvent(ev *tcell.EventKey) (outbound.InputEvent, bool) {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC, tcell.KeyCtrlZ:
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
