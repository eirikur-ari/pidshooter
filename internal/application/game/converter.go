package game

import (
	"math"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

func toConfirmViewState(target *game.Target) *outbound.ConfirmViewState {
	if target == nil {
		return nil
	}
	return &outbound.ConfirmViewState{Pid: target.Pid, Name: target.Name}
}

func toTargetViewState(target *game.Target) outbound.TargetViewState {
	return outbound.TargetViewState{
		X:       int(math.Round(target.Position.X)),
		Y:       int(math.Round(target.Position.Y)),
		Tag:     target.Tag(),
		Killing: target.State == game.Killing,
	}
}

func toTargetViewStates(targets []*game.Target) []outbound.TargetViewState {
	views := make([]outbound.TargetViewState, len(targets))
	for i, target := range targets {
		views[i] = toTargetViewState(target)
	}
	return views
}

func toHUDState(tracker *killTracker) outbound.HUDState {
	return outbound.HUDState{
		FreedMem:  tracker.score.freedMem,
		Kills:     tracker.score.kills,
		HighScore: tracker.score.highScore,
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

func toFrameState(session *game.Session, tracker *killTracker) outbound.FrameState {
	targets, alive := session.AvailableTargets()

	return outbound.FrameState{
		Targets:   toTargetViewStates(targets),
		HUD:       toHUDState(tracker),
		StatusBar: toStatusState(session, alive),
	}
}
