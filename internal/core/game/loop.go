package game

import "github.com/eirikur-ari/pidshooter/internal/core/movement"

// Update advances the game state by one tick. w and h are the current terminal dimensions.
func (g *Game) Update(w, h int) {
	if g.timer.Expired() {
		g.Stop()
		return
	}

	g.moveOrDie(w, h)

	if g.allTargetsDead() {
		g.Stop()
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
	bounds := movement.NewBounds(w, h)
	speed := g.throttle.Speed()
	for _, t := range g.targets {
		t.Update(bounds, speed)
	}
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
