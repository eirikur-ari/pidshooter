package cli

// FakeRunner is a test double for inbound.Runner. It returns Err.
type FakeRunner struct {
	Err error
}

func (r *FakeRunner) Run() error {
	return r.Err
}
