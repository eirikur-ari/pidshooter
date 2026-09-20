package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"

// RunnerFactory is a test double for inbound.RunnerFactory. It returns
// Runner and Err, and records how many times Create was called.
type RunnerFactory struct {
	Runner inbound.Runner
	Err    error
	Calls  int
}

func (f *RunnerFactory) Create() (inbound.Runner, error) {
	f.Calls++
	return f.Runner, f.Err
}
