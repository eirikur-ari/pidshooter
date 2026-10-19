package inbound

// Runner runs what a request describes.
type Runner interface {
	// Run runs what request describes, returning an error if it could not
	// complete.
	Run(request RunRequest) error
}

// RunRequest describes one thing to run. Exactly one field is set.
type RunRequest struct {
	Play *PlayRequest
	List *ListRequest
	Kill *KillRequest
}
