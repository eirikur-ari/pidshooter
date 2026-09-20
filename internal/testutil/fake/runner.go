package fake

import "github.com/eirikur-ari/pidshooter/internal/application/config"

// Runner is a test double for inbound.Runner. It captures the Config it was
// run with and returns Err.
type Runner struct {
	Err error
	Cfg config.Config
}

func (r *Runner) Run(cfg config.Config) error {
	r.Cfg = cfg
	return r.Err
}
