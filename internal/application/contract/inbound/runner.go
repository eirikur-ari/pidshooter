package inbound

// Runner performs one unit of work on request.
type Runner interface {
	// Run performs the work, returning an error if it could not complete.
	Run() error
}
