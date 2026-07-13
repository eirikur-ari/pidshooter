package game

import (
	"time"
)

// Update advances the game state by one tick. w and h are the current terminal dimensions.
func (g *Game) Update(w, h int) {
	if g.timeExpired() {
		g.Stop()
		return
	}

	g.moveOrDie(w, h)

	if g.allTargetsDead() {
		g.Stop()
	}
}

// Frame returns a snapshot of current game state for the renderer.
func (g *Game) Frame() FrameState {
	targets, alive := g.targetViews()

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
			TimeLeft:   g.timeLeftSeconds(),
			Confirming: g.confirm.ViewState(),
		},
	}
}

// Kill is called by the application layer after a successful OS kill.
// It transitions the target to the kill animation and records the session stats.
func (g *Game) Kill(t *Target) {
	if t.Kill() {
		g.RecordKill(t.Rss)
	}
}

func (g *Game) moveOrDie(w, h int) {
	bounds := FrameBounds{Width: w, Height: h}
	speed := g.velocity.Speed()
	for _, t := range g.targets {
		t.Update(bounds, speed)
	}
}

func (g *Game) targetViews() (views []TargetViewState, alive int) {
	views = make([]TargetViewState, 0, len(g.targets))
	for _, t := range g.targets {
		if t.IsDead() {
			continue
		}
		if t.IsAlive() {
			alive++
		}
		views = append(views, t.ViewState())
	}
	return views, alive
}

func (g *Game) timeLeftSeconds() int {
	if g.cfg.TimeLimit <= 0 {
		return 0
	}
	return int(g.timeRemaining().Seconds())
}

func (g *Game) timeExpired() bool {
	return g.cfg.TimeLimit > 0 && g.timeRemaining() == 0
}

func (g *Game) allTargetsDead() bool {
	// No targets mean nothing was ever killed, not that everything was.
	if len(g.targets) == 0 {
		return false
	}
	for _, t := range g.targets {
		if !t.IsDead() {
			return false
		}
	}
	return true
}

func (g *Game) timeRemaining() time.Duration {
	remaining := time.Duration(g.cfg.TimeLimit)*time.Second - time.Since(g.startTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}
