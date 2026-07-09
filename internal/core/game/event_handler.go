package game

// HandleClick processes a mouse click and returns a KillRequest if a target is hit.
func (g *Game) HandleClick(x, y int) *KillRequest {
	if g.confirming != nil {
		return nil
	}
	for _, e := range g.targets {
		if e.Contains(x, y) {
			if g.confirmMode {
				g.confirming = e
			} else {
				return &KillRequest{Target: e}
			}
			return nil
		}
	}
	return nil
}

// HandleKey processes a rune key press and returns a KillRequest if a pending
// confirmation is accepted. Control keys (Escape, CtrlC, CtrlZ) are handled by
// the application layer before this is called.
func (g *Game) HandleKey(ch rune) *KillRequest {
	switch {
	case ch == 'q' || ch == 'Q':
		if g.confirming != nil {
			g.confirming = nil
		} else {
			g.running.Store(false)
		}
	case ch == 'y' || ch == 'Y':
		if g.confirming != nil {
			target := g.confirming
			g.confirming = nil
			return &KillRequest{Target: target}
		}
	case ch == 'n' || ch == 'N':
		if g.confirming != nil {
			g.confirming = nil
		}
	case ch == '+' || ch == '=':
		g.speed += 0.5
		if g.speed > 5.0 {
			g.speed = 5.0
		}
	case ch == '-' || ch == '_':
		g.speed -= 0.5
		if g.speed < 0.1 {
			g.speed = 0.1
		}
	}
	return nil
}
