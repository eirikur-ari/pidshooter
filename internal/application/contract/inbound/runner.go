package inbound

// Runner is the inbound port: the contract the delivery adapter calls.
type Runner interface {
	// Run executes with the given configuration, returning any error encountered.
	Run(cfg Config) error
}
