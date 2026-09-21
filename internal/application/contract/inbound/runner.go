package inbound

import "github.com/eirikur-ari/pidshooter/internal/application/config"

// Runner is an inbound port: the contract an inbound adapter calls.
type Runner interface {
	// Run executes with the given configuration, returning an error when something happens.
	Run(cfg config.Config) error
}
