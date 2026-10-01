package inbound

// Runner is an inbound port: the contract an inbound adapter calls.
type Runner interface {
	// Run performs the inbound adapter's requested work, returning an error
	// if it could not complete.
	Run() error
}
