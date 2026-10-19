package inbound

// KillRunner kills the processes that match search patterns.
type KillRunner interface {
	// Run kills the processes described by request, returning an error if it
	// could not complete.
	Run(request KillRequest) error
}

// KillRequest describes the processes to kill.
type KillRequest struct {
	// Patterns select the processes to kill.
	Patterns []string
}
