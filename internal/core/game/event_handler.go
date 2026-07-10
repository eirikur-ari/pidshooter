package game

// HandleClick processes a mouse click and returns the target to kill if one is hit.
func (g *Game) HandleClick(x, y int) *Target {
	if g.confirming != nil {
		return nil
	}
	for _, e := range g.targets {
		if e.Contains(x, y) {
			if g.cfg.ConfirmMode {
				g.confirming = e
				return nil
			}
			return e
		}
	}
	return nil
}

// HandleKey processes a rune key press and returns the target to kill if a pending
// confirmation is accepted. Control keys (Escape, CtrlC, CtrlZ) are handled by
// the application layer before this is called.
func (g *Game) HandleKey(ch rune) *Target {
	switch {
	case ch == 'q' || ch == 'Q':
		if g.confirming != nil {
			g.confirming = nil
		} else {
			g.Stop()
		}
	case ch == 'y' || ch == 'Y':
		if g.confirming != nil {
			target := g.confirming
			g.confirming = nil
			return target
		}
	case ch == 'n' || ch == 'N':
		if g.confirming != nil {
			g.confirming = nil
		}
	case ch == '+' || ch == '=':
		g.velocity.Increase()
	case ch == '-' || ch == '_':
		g.velocity.Decrease()
	}
	return nil
}

