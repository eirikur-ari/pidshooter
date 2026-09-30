package fake

// Runner is a test double for inbound.Runner. It returns Err.
type Runner struct {
	Err error
}

func (r *Runner) Run() error {
	return r.Err
}
