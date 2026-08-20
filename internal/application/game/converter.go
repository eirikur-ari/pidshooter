package game

import (
	"math"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
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

func toHUDState(tracker *score.Tracker) outbound.HUDState {
	return outbound.HUDState{
		FreedMem:  tracker.FreedMem,
		Kills:     tracker.Kills,
		HighScore: tracker.HighScore,
	}
}

func toStatusState(s *game.Session, alive int) outbound.StatusState {
	return outbound.StatusState{
		Alive:      alive,
		Speed:      s.Throttle().Speed(),
		TimeLimit:  s.TimeLimit(),
		TimeLeft:   s.TimeLeft(),
		Confirming: toConfirmViewState(s.PendingConfirm()),
	}
}
