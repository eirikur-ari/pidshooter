package inbound

// PlayRunner plays a game session.
type PlayRunner interface {
	// Run plays a game session as described by request, returning an error if
	// it could not complete.
	Run(request PlayRequest) error
}

// ListRunner lists the processes that match search patterns.
type ListRunner interface {
	// Run lists the processes described by request, returning an error if it
	// could not complete.
	Run(request ListRequest) error
}

// KillRunner kills the processes that match search patterns.
type KillRunner interface {
	// Run kills the processes described by request, returning an error if it
	// could not complete.
	Run(request KillRequest) error
}

// PlayRequest describes a game session to play.
type PlayRequest struct {
	// Patterns select the processes to target.
	Patterns []string
	Game     GameOptions
	Process  ProcessOptions
}

// ListRequest describes the processes to list.
type ListRequest struct {
	// Patterns select the processes to list.
	Patterns []string
	Process  ProcessOptions
}

// KillRequest describes the processes to kill.
type KillRequest struct {
	// Patterns select the processes to kill.
	Patterns []string
	Process  ProcessOptions
}

// GameOptions holds explicitly provided game values. A nil field means the
// value was not provided.
type GameOptions struct {
	ConfirmMode *bool
	Speed       *float64
	TimeLimit   *int
}

// ProcessOptions holds explicitly provided process-discovery values. A nil
// field means the value was not provided.
type ProcessOptions struct {
	// IncludeRoot additionally permits root-owned processes as targets.
	IncludeRoot *bool
	// AllowRoot permits running as root.
	AllowRoot *bool
}
