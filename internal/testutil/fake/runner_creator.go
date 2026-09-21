package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"

// RunnerCreator is a test double that constructs a Runner. It returns
// Runner and Err, and records how many times Create was called.
type RunnerCreator struct {
	Runner inbound.Runner
	Err    error
	Calls  int
}

// Create returns Runner and Err, and records the call.
func (f *RunnerCreator) Create() (inbound.Runner, error) {
	f.Calls++
	return f.Runner, f.Err
}
