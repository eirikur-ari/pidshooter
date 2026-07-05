package game

import (
	"time"

	"github.com/eirikur-ari/pidshooter/internal/domain/process"
	gamedriven "github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"
)

const (
	targetFPS     = 20
	frameDuration = time.Second / targetFPS
)

func (g *Game) populateEntities(processes []process.Info, w, h int) {
	for _, p := range processes {
		g.targets = append(g.targets, NewTarget(p, w, h))
	}
}

func (g *Game) run() {
	g.startTime = time.Now()
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	for g.running.Load() {
		g.drainEvents()
		g.update()
		g.render()
		<-ticker.C
	}
}

func (g *Game) update() {
	w, h := g.renderer.Size()

	if g.timeLimit > 0 && g.timeRemaining() == 0 {
		g.running.Store(false)
		return
	}

	allDead := true
	for _, e := range g.targets {
		e.Update(w, h, g.speed)
		if e.State != Dead {
			allDead = false
		}
	}

	if allDead && len(g.targets) > 0 {
		g.running.Store(false)
	}
}

func (g *Game) render() {
	targets := make([]gamedriven.TargetView, 0, len(g.targets))
	for _, e := range g.targets {
		if e.State == Dead {
			continue
		}
		targets = append(targets, gamedriven.TargetView{
			X:       int(e.PosX),
			Y:       int(e.PosY),
			Label:   e.Label(),
			Killing: e.State == Killing,
		})
	}

	alive := 0
	for _, e := range g.targets {
		if e.State == Alive {
			alive++
		}
	}

	var cs *gamedriven.ConfirmState
	if g.confirming != nil {
		cs = &gamedriven.ConfirmState{PID: g.confirming.Pid, Name: g.confirming.Name}
	}

	var timeLeft int
	if g.timeLimit > 0 {
		timeLeft = int(g.timeRemaining().Seconds())
	}

	g.renderer.Render(gamedriven.Frame{
		Targets: targets,
		HUD: gamedriven.HUDState{
			FreedMem:  g.freedMem,
			Kills:     g.kills,
			HighScore: g.highScore,
		},
		StatusBar: gamedriven.StatusState{
			Alive:      alive,
			Speed:      g.speed,
			TimeLimit:  g.timeLimit,
			TimeLeft:   timeLeft,
			Confirming: cs,
		},
	})
}

func (g *Game) killTarget(e *Target) {
	if e.State != Alive {
		return
	}
	if err := g.killer.Kill(e.Pid, e.Name); err != nil {
		return
	}
	e.StartKillAnim()
	g.Session.RecordKill(e.Rss)
}

func (g *Game) timeRemaining() time.Duration {
	remaining := time.Duration(g.timeLimit)*time.Second - time.Since(g.startTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}
