package game

// HandleClick processes a click at game-space coordinates (x=column, y=row for
// terminal UIs; translate to game-space before calling for other renderers) and
// returns the target to kill if one is hit.
func (g *Game) HandleClick(x, y int) *Target {
	if g.confirm.Pending() {
		return nil
	}
	for _, t := range g.targets {
		if t.IsHitAt(x, y) {
			return g.confirm.Request(t)
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
		return g.confirm.Accept()
	case ch == 'n' || ch == 'N':
		g.confirm.Cancel()
	case ch == '+' || ch == '=':
		g.velocity.Increase()
	case ch == '-' || ch == '_':
		g.velocity.Decrease()
	}
	return nil
}

// onQuit cancels a pending confirmation if one exists, otherwise stops the game.
func (g *Game) onQuit() {
	if g.confirm.Pending() {
		g.confirm.Cancel()
	} else {
		g.Stop()
	}
}
