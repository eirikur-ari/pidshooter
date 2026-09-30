package inbound

// Runner is an inbound port: the contract an inbound adapter calls.
type Runner interface {
	// Run executes, returning an error when something happens.
	Run() error
}
