package game

import (
	"math"
	"time"
)

// Update advances the game state by one tick. w and h are the current terminal dimensions.
func (g *Game) Update(w, h int) {
	if g.cfg.TimeLimit > 0 && g.timeRemaining() == 0 {
		g.Stop()
		return
	}

	bounds := FrameBounds{W: w, H: h}
	allDead := true
	for _, t := range g.targets {
		t.Update(bounds, g.velocity.Speed())
		if t.State != Dead {
			allDead = false
		}
	}

	if allDead && len(g.targets) > 0 {
		g.Stop()
	}
}

// Frame returns a snapshot of current game state for the renderer.
func (g *Game) Frame() FrameState {
	targets := make([]TargetViewState, 0, len(g.targets))
	alive := 0
	for _, t := range g.targets {
		if t.State == Dead {
			continue
		}
		if t.State == Alive {
			alive++
		}
		targets = append(targets, TargetViewState{
			X:       int(math.Round(t.Position.X)),
			Y:       int(math.Round(t.Position.Y)),
			Tag:     t.Tag(),
			Killing: t.State == Killing,
		})
	}

	cs := g.confirm.View()

	var timeLeft int
	if g.cfg.TimeLimit > 0 {
		timeLeft = int(g.timeRemaining().Seconds())
	}

	return FrameState{
		Targets: targets,
		HUD: HUDState{
			FreedMem:  g.freedMem,
			Kills:     g.kills,
			HighScore: g.highScore,
		},
		StatusBar: StatusState{
			Alive:      alive,
			Speed:      g.velocity.Speed(),
			TimeLimit:  g.cfg.TimeLimit,
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
	t.StartKillAnimation()
	g.Session.RecordKill(t.Rss)
}

func (g *Game) timeRemaining() time.Duration {
	remaining := time.Duration(g.cfg.TimeLimit)*time.Second - time.Since(g.startTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}
