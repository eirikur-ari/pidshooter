package outbound

// FrameState is the data snapshot passed to the Renderer each tick.
type FrameState struct {
	Targets   []TargetViewState
	HUD       HUDState
	StatusBar StatusState
}

// TargetViewState is the render representation of a single target.
type TargetViewState struct {
	// X, Y is the target's position, in the same character-cell coordinate
	// space as Renderer.Size, with the origin at the top-left.
	X, Y int
	// Tag is the label drawn at X, Y, rightward.
	Tag string
	// Killing reports whether the target's kill animation is playing.
	Killing bool
}

// HUDState carries the heads-up display values.
type HUDState struct {
	// FreedMem is the cumulative memory freed by kills, in bytes.
	FreedMem  int64
	Kills     int
	HighScore int
}

// StatusState carries the status bar values.
type StatusState struct {
	// Alive is the number of targets not yet killed.
	Alive int
	// Speed is the current target movement speed, as a multiplier of the
	// base speed.
	Speed float64
	// TimeLimit is the session's time limit in seconds, or 0 if untimed.
	TimeLimit int
	// TimeLeft is the seconds remaining in the session. Meaningful only
	// when TimeLimit > 0.
	TimeLeft int
	// Confirming holds the kill-confirmation prompt values, or nil when no
	// confirmation is pending.
	Confirming *ConfirmViewState
}

// ConfirmViewState carries the kill-confirmation prompt values.
type ConfirmViewState struct {
	Pid  int
	Name string
}

// Renderer is the outbound port for terminal rendering.
//
// Init must be called and succeed before Size or Render is used. Cleanup
// releases what Init acquired and is safe to call more than once.
type Renderer interface {
	// Init prepares the terminal for rendering and input.
	Init() error
	// Cleanup releases what Init acquired, restoring the terminal to its
	// prior state. Idempotent: calls after the first are no-ops.
	Cleanup()
	// Size returns the current terminal dimensions, in character cells.
	// It may change between calls, e.g. if the terminal is resized.
	Size() (width, height int)
	// Render draws state to the terminal, replacing whatever was
	// previously rendered.
	Render(state FrameState)
}
