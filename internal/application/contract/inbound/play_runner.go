package inbound

// PlayRunner plays a game session.
type PlayRunner interface {
	// Run plays a game session as described by request, returning an error if
	// it could not complete.
	Run(request PlayRequest) error
}

// PlayRequest describes a game session to play.
type PlayRequest struct {
	// Patterns select the processes to target.
	Patterns []string
	Game     GameOptions
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
	IncludeRoot *bool
	// AllowRoot permits running as root.
	AllowRoot *bool
}
