package inbound

import "github.com/eirikur-ari/pidshooter/internal/application/config"

// RunRequest holds the inbound adapter's mapped input values (e.g. CLI
// flags). Patterns has no default and is always taken from the caller.
type RunRequest struct {
	Patterns []string
	Config   config.Request
}

// Runner is an inbound port: the contract an inbound adapter calls.
type Runner interface {
	// Run executes with the given request, returning an error when something happens.
	Run(req RunRequest) error
}
