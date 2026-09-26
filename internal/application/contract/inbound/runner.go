package inbound

// GameConfig holds the caller's explicitly-provided game mode values. A nil
// field means the caller did not provide that value.
type GameConfig struct {
	ConfirmMode *bool
	Speed       *float64
	TimeLimit   *int
}

// ProcessConfig holds the caller's explicitly-provided process-discovery
// values. A nil field means the caller did not provide that value.
type ProcessConfig struct {
	IncludeRoot *bool
}

// RunRequest holds the inbound adapter's mapped input values (e.g. CLI
// flags). Patterns has no default and is always taken from the caller.
type RunRequest struct {
	Patterns []string
	Game     GameConfig
	Process  ProcessConfig
}

// Runner is an inbound port: the contract an inbound adapter calls.
type Runner interface {
	// Run executes with the given request, returning an error when something happens.
	Run(req RunRequest) error
}
