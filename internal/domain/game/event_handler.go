package game

import "github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"

func (g *Game) drainEvents() {
	for {
		select {
		case ev := <-g.events.Events():
			g.handleEvent(ev)
		default:
			return
		}
	}
}

func (g *Game) handleEvent(ev driven.InputEvent) {
	switch ev := ev.(type) {
	case driven.ClickEvent:
		g.handleMouseClick(ev.X, ev.Y)
	case driven.KeyEvent:
		g.handleKeyPress(ev.Key, ev.Ch)
	case driven.ResizeEvent:
		// size is re-queried from the renderer each tick
	}
}

func (g *Game) handleMouseClick(x, y int) {
	if g.confirming != nil {
		return
	}

	for _, e := range g.targets {
		if e.Contains(x, y) {
			if g.confirmMode {
				g.confirming = e
			} else {
				g.killTarget(e)
			}
			return
		}
	}
}

func (g *Game) handleKeyPress(key driven.KeyCode, r rune) {
	switch {
	case key == driven.KeyEscape:
		g.running.Store(false)
	case key == driven.KeyCtrlC:
		g.running.Store(false)
	case key == driven.KeyCtrlZ:
		g.running.Store(false)
	case r == 'q' || r == 'Q':
		if g.confirming != nil {
			g.confirming = nil
		} else {
			g.running.Store(false)
		}
	case r == 'y' || r == 'Y':
		if g.confirming != nil {
			g.killTarget(g.confirming)
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
