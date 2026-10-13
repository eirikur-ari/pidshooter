package outbound

// FrameViewState is everything to draw in a single frame.
type FrameViewState struct {
	Targets   []TargetViewState
	HUD       HUDViewState
	StatusBar StatusViewState
}

// TargetViewState describes how to draw a single target.
type TargetViewState struct {
	// X, Y is the target's position within the renderer's window, with the
	// origin at the top-left.
	X, Y int
	// Tag is the label drawn at X, Y, rightward. Meaningful only when
	// neither Killing nor Fleeing is true; an animation frame is drawn
	// instead of Tag while either is playing.
	Tag string
	// Killing reports whether the target's kill animation is playing.
	Killing bool
	// Fleeing reports whether the target's flee animation is playing.
	Fleeing bool
	// AnimationProgress is how far through its kill or flee animation the
	// target is, from 0 to 1. Meaningful only when Killing or Fleeing is true.
	AnimationProgress float64
}

// HUDViewState holds the values shown in the heads-up display.
type HUDViewState struct {
	// FreedMem is the cumulative memory freed by kills, in bytes.
	FreedMem int64
	// Kills is the number of kills so far in the session.
	Kills int
	// HighScore is the kill count to beat for a new high score.
	HighScore int
}

// StatusViewState holds the values shown in the status bar.
type StatusViewState struct {
	// Alive is the number of targets not yet killed.
	Alive int
	// Speed is the current target movement speed, as a multiplier of the
	// base speed.
	Speed float64
	// TimeLeft is the seconds remaining in the session, or nil if untimed.
	TimeLeft *int
	// Confirming holds the kill-confirmation prompt values, or nil when no
	// confirmation is pending.
	Confirming *ConfirmViewState
}

// ConfirmViewState holds the values shown in the kill-confirmation prompt.
type ConfirmViewState struct {
	PID  int
	Name string
}

// WindowSize is the width and height of the display.
type WindowSize struct {
	Width, Height int
}

// ChromeSize is the amount of space reserved at the top and bottom edges of
// the display for a renderer's own fixed UI, in the same units as
// WindowSize.
type ChromeSize struct {
	Top, Bottom int
}

// Renderer draws frames to the display. WindowSize and Render are valid
// only after a successful Init.
type Renderer interface {
	// Init prepares the display for rendering.
	Init() error
	// Cleanup releases what Init acquired, restoring the display to its
	// prior state.
	Cleanup()
	// WindowSize returns the current display dimensions. It may change
	// between calls, e.g. if the display is resized.
	WindowSize() WindowSize
	// ChromeSize returns the space this renderer reserves for its own fixed
	// UI. Constant for the renderer's lifetime.
	ChromeSize() ChromeSize
	// Render draws state to the display, replacing whatever was
	// previously rendered.
	Render(state FrameViewState)
}
