package inbound

// ListRunner lists the processes that match search patterns.
type ListRunner interface {
	// Run lists the processes described by request, returning an error if it
	// could not complete.
	Run(request ListRequest) error
}

// ListRequest describes the processes to list.
type ListRequest struct {
	// Patterns select the processes to list.
	Patterns []string
}
