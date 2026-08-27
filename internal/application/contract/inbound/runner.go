package inbound

// Runner is the inbound port: the contract the delivery adapter calls.
type Runner interface {
	// Run executes with the given configuration, returning an error only
	// if it could not complete.
	Run(cfg Config) error
}
