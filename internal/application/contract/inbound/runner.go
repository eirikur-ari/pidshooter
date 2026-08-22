package inbound

// Runner is the inbound port: the contract the delivery adapter calls.
type Runner interface {
	Run(cfg Config) error
}
