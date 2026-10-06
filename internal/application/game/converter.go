package game

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
)

// toGameConfig converts a PlayRequest to a game.Config.
func toGameConfig(req PlayRequest) game.Config {
	return game.Config{Confirm: req.ConfirmMode, Speed: req.Speed, TimeLimit: req.TimeLimit}
}

// toWindowSize converts an outbound.WindowSize to a movement.WindowSize.
func toWindowSize(window outbound.WindowSize) movement.WindowSize {
	return movement.WindowSize{Width: window.Width, Height: window.Height}
}

// toBounds converts a window size and chrome size to a movement.Bounds.
func toBounds(window outbound.WindowSize, chrome outbound.ChromeSize) movement.Bounds {
	return movement.NewBounds(
		toWindowSize(window),
		movement.ChromeSize{Top: chrome.Top, Bottom: chrome.Bottom},
	)
}

// toConfirmViewState converts a target to a ConfirmViewState, or nil if there is none.
func toConfirmViewState(target *game.Target) *outbound.ConfirmViewState {
	if target == nil {
		return nil
	}
	return &outbound.ConfirmViewState{PID: target.Info.PID, Name: target.Info.Name}
}

// toTargetViewState converts a target to a TargetViewState, or an empty one if there is none.
func toTargetViewState(target *game.Target) outbound.TargetViewState {
	if target == nil {
		return outbound.TargetViewState{}
	}
	x, y := target.Motion.Position.Rounded()
	return outbound.TargetViewState{
		X:                 x,
		Y:                 y,
		Tag:               target.Tag(),
		Killing:           target.State == game.Killing,
		Fleeing:           target.State == game.Fleeing,
		AnimationProgress: target.AnimationProgress(),
	}
}

// toTargetViewStates converts targets to TargetViewStates, keeping their order.
func toTargetViewStates(targets []*game.Target) []outbound.TargetViewState {
	views := make([]outbound.TargetViewState, len(targets))
	for i, target := range targets {
		views[i] = toTargetViewState(target)
	}
	return views
}

// toHUDViewState converts the tracker's progress to a HUDViewState, or an empty one if there is none.
func toHUDViewState(tracker *killTracker) outbound.HUDViewState {
	if tracker == nil {
		return outbound.HUDViewState{}
	}
	return outbound.HUDViewState{
		FreedMem:  tracker.freedMem(),
		Kills:     tracker.kills(),
		HighScore: tracker.highScore(),
	}
}

// toStatusViewState converts the session state and alive count to a StatusViewState, or an empty one if there is no session.
func toStatusViewState(session *game.Session, alive int) outbound.StatusViewState {
	if session == nil {
		return outbound.StatusViewState{}
	}
	var timeLeft *int
	if session.TimeLimit() > 0 {
		remaining := session.TimeLeft()
		timeLeft = &remaining
	}
	return outbound.StatusViewState{
		Alive:      alive,
		Speed:      session.Throttle().Speed(),
		TimeLeft:   timeLeft,
		Confirming: toConfirmViewState(session.PendingConfirm()),
	}
}

// toFrameViewState converts the session and tracker to a FrameViewState.
// A missing session or tracker leaves its part of the frame empty.
func toFrameViewState(session *game.Session, tracker *killTracker) outbound.FrameViewState {
	var targets []*game.Target
	alive := 0
	if session != nil {
		targets, alive = session.AvailableTargets()
	}

	return outbound.FrameViewState{
		Targets:   toTargetViewStates(targets),
		HUD:       toHUDViewState(tracker),
		StatusBar: toStatusViewState(session, alive),
	}
}
