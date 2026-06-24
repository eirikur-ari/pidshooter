package game

import "github.com/gdamore/tcell/v2"

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
		g.handleKeyPress(ev.Key(), ev.Rune())
	case *tcell.EventResize:
		g.screen.Sync()
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
			g.confirming = nil
		} else {
			g.running = false
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