package game

import (
	"math"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

func toConfirmViewState(t *game.Target) *outbound.ConfirmViewState {
	if t == nil {
		return nil
	}
	return &outbound.ConfirmViewState{PID: t.Pid, Name: t.Name}
}

func toTargetViewState(t *game.Target) outbound.TargetViewState {
	return outbound.TargetViewState{
		X:       int(math.Round(t.Position.X)),
		Y:       int(math.Round(t.Position.Y)),
		Tag:     t.Tag(),
		Killing: t.State == game.Killing,
	}
}

func toTargetViewStates(targets []*game.Target) []outbound.TargetViewState {
	views := make([]outbound.TargetViewState, len(targets))
	for i, t := range targets {
		views[i] = toTargetViewState(t)
	}
	return views
}

func toHUDState(tracker *scoreTracker) outbound.HUDState {
	return outbound.HUDState{
		FreedMem:  tracker.freedMem,
		Kills:     tracker.kills,
		HighScore: tracker.highScore,
	}
}

func toStatusState(session *game.Session, alive int) outbound.StatusState {
	return outbound.StatusState{
		Alive:      alive,
		Speed:      session.Throttle().Speed(),
		TimeLimit:  session.TimeLimit(),
		TimeLeft:   session.TimeLeft(),
		Confirming: toConfirmViewState(session.PendingConfirm()),
	}
}

func toFrameState(session *game.Session, tracker *scoreTracker) outbound.FrameState {
	targets, alive := session.AvailableTargets()

	return outbound.FrameState{
		Targets:   toTargetViewStates(targets),
		HUD:       toHUDState(tracker),
		StatusBar: toStatusState(session, alive),
	}
}
