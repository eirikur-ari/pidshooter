package game

import "time"

// Update advances the game state by one tick. w and h are the current terminal dimensions.
func (g *Game) Update(w, h int) {
	if g.timeLimit > 0 && g.timeRemaining() == 0 {
		g.running.Store(false)
		return
	}

	allDead := true
	for _, e := range g.targets {
		e.Update(w, h, g.velocity.Speed())
		if e.State != Dead {
			allDead = false
		}
	}

	if allDead && len(g.targets) > 0 {
		g.running.Store(false)
	}
}

// Frame returns a snapshot of current game state for the renderer.
func (g *Game) Frame() Frame {
	targets := make([]TargetView, 0, len(g.targets))
	alive := 0
	for _, e := range g.targets {
		if e.State == Dead {
			continue
		}
		if e.State == Alive {
			alive++
		}
		targets = append(targets, TargetView{
			X:       int(e.Position.X),
			Y:       int(e.Position.Y),
			Label:   e.Label(),
			Killing: e.State == Killing,
		})
	}

	var cs *ConfirmState
	if g.confirming != nil {
		cs = &ConfirmState{PID: g.confirming.Pid, Name: g.confirming.Name}
	}

	var timeLeft int
	if g.timeLimit > 0 {
		timeLeft = int(g.timeRemaining().Seconds())
	}

	return Frame{
		Targets: targets,
		HUD: HUDState{
			FreedMem:  g.freedMem,
			Kills:     g.kills,
			HighScore: g.highScore,
		},
		StatusBar: StatusState{
			Alive:      alive,
			Speed:      g.velocity.Speed(),
			TimeLimit:  g.timeLimit,
			TimeLeft:   timeLeft,
			Confirming: cs,
		},
	}
}

// CompleteKill is called by the application layer after a successful OS kill.
// It transitions the target to the kill animation and records the session stats.
func (g *Game) CompleteKill(t *Target) {
	if t.State != Alive {
		return
	}
	t.StartKillAnim()
	g.Session.RecordKill(t.Rss)
}

func (g *Game) timeRemaining() time.Duration {
	remaining := time.Duration(g.timeLimit)*time.Second - time.Since(g.startTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}
