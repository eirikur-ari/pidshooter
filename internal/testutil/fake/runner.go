package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"

// Runner is a test double for inbound.Runner. It captures the RunRequest it
// was run with and returns Err.
type Runner struct {
	Err error
	Req inbound.RunRequest
}

func (r *Runner) Run(req inbound.RunRequest) error {
	r.Req = req
	return r.Err
}
