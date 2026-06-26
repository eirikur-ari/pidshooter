package driven

// Renderer is the driven port for terminal rendering.
type Renderer interface {
	Render(frame Frame)
	Size() (width, height int)
	Init() error
	Cleanup()
}

// Frame is the data snapshot the domain passes to Renderer each tick.
type Frame struct {
	Targets   []TargetView
	HUD       HUDState
	StatusBar StatusState
}

// TargetView is the render representation of a single target.
type TargetView struct {
	X, Y    int
	Label   string
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
	Confirming *ConfirmState
}

// ConfirmState carries the kill-confirmation prompt values.
type ConfirmState struct {
	PID  int
	Name string
}
