package outbound

// FrameState is the data snapshot passed to the Renderer each tick.
type FrameState struct {
	Targets   []TargetViewState
	HUD       HUDState
	StatusBar StatusState
}

// TargetViewState is the render representation of a single target.
type TargetViewState struct {
	X, Y    int
	Tag     string
	Killing bool
}

// HUDState carries the heads-up display values.
type HUDState struct {
	FreedMem  int64
	Kills     int
	HighScore int
}

// StatusState carries the status bar values.
type StatusState struct {
	Alive      int
	Speed      float64
	TimeLimit  int
	TimeLeft   int
	Confirming *ConfirmViewState
}

// ConfirmViewState carries the kill-confirmation prompt values.
type ConfirmViewState struct {
	PID  int
	Name string
}

// Renderer is the outbound port for terminal rendering.
type Renderer interface {
	Init() error
	Cleanup()
	Size() (width, height int)
	Render(state FrameState)
}
