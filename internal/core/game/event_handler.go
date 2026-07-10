package game

// HandleClick processes a click at game-space coordinates (x=column, y=row for
// terminal UIs; translate to game-space before calling for other renderers) and
// returns the target to kill if one is hit.
func (g *Game) HandleClick(x, y int) *Target {
	if g.confirm.Pending() {
		return nil
	}
	for _, t := range g.targets {
		if t.Contains(x, y) {
			if g.cfg.ConfirmMode {
				g.confirm.Set(t)
				return nil
			}
			return t
		}
	}
	return nil
}

// HandleKey processes a printable character key press (ch is the Unicode rune,
// matching KeyEvent.Ch from the event layer). Returns the target to kill if a
// pending confirmation is accepted. Control keys (Escape, CtrlC, CtrlZ) are
// handled by the application layer before this is called.
func (g *Game) HandleKey(ch rune) *Target {
	switch {
	case ch == 'q' || ch == 'Q':
		g.onQuit()
	case ch == 'y' || ch == 'Y':
		return g.confirm.Kill()
	case ch == 'n' || ch == 'N':
		g.confirm.Clear()
	case ch == '+' || ch == '=':
		g.velocity.Increase()
	case ch == '-' || ch == '_':
		g.velocity.Decrease()
	}
	return nil
}

func (g *Game) onQuit() {
	if g.confirm.Pending() {
		g.confirm.Clear()
	} else {
		g.Stop()
	}
}
