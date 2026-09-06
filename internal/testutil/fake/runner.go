package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"

// Runner is a test double for inbound.Runner. It captures the Config it was
// run with and returns Err.
type Runner struct {
	Err error
	Cfg inbound.Config
}

func (r *Runner) Run(cfg inbound.Config) error {
	r.Cfg = cfg
	return r.Err
}