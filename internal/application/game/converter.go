package game

import (
	"math"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
)

// toGameConfig converts a PlayRequest into the game.Config.
func toGameConfig(req PlayRequest) game.Config {
	return game.Config{Confirm: req.ConfirmMode, Speed: req.Speed, TimeLimit: req.TimeLimit}
}

// toBounds converts the display dimensions and reserved chrome rows
// reported by a Renderer into a movement.Bounds.
func toBounds(window outbound.WindowSize, chrome outbound.ChromeSize) movement.Bounds {
	return movement.NewBounds(
		movement.WindowSize{Width: window.Width, Height: window.Height},
		movement.ChromeSize{Top: chrome.Top, Bottom: chrome.Bottom},
	)
}

func toConfirmViewState(target *game.Target) *outbound.ConfirmViewState {
	if target == nil {
		return nil
	}
	return &outbound.ConfirmViewState{PID: target.Info.PID, Name: target.Info.Name}
}

func toTargetViewState(target *game.Target) outbound.TargetViewState {
	return outbound.TargetViewState{
		X:                 int(math.Round(target.Motion.Position.X)),
		Y:                 int(math.Round(target.Motion.Position.Y)),
		Tag:               target.Tag(),
		Killing:           target.State == game.Killing,
		Fleeing:           target.State == game.Fleeing,
		AnimationProgress: target.AnimationProgress(),
	}
}

func toTargetViewStates(targets []*game.Target) []outbound.TargetViewState {
	views := make([]outbound.TargetViewState, len(targets))
	for i, target := range targets {
		views[i] = toTargetViewState(target)
	}
	return views
}

func toHUDViewState(tracker *killTracker) outbound.HUDViewState {
	return outbound.HUDViewState{
		FreedMem:  tracker.score.freedMem,
		Kills:     tracker.score.kills,
		HighScore: tracker.score.highScore,
	}
}

func toStatusViewState(session *game.Session, alive int) outbound.StatusViewState {
	var timeLeft *int
	if session.TimeLimit() > 0 {
		t := session.TimeLeft()
		timeLeft = &t
	}
	return outbound.StatusViewState{
		Alive:      alive,
		Speed:      session.Throttle().Speed(),
		TimeLeft:   timeLeft,
		Confirming: toConfirmViewState(session.PendingConfirm()),
	}
}

func toFrameViewState(session *game.Session, tracker *killTracker) outbound.FrameViewState {
	targets, alive := session.AvailableTargets()

	return outbound.FrameViewState{
		Targets:   toTargetViewStates(targets),
		HUD:       toHUDViewState(tracker),
		StatusBar: toStatusViewState(session, alive),
	}
}
