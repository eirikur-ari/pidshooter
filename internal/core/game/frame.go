package game

// FrameState is the data snapshot the application layer passes to the Renderer each tick.
type FrameState struct {
	Targets   []TargetViewState
	HUD       HUDState
	StatusBar StatusState
}

// TargetViewState is the render representation of a single target.
type TargetViewState struct {
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
