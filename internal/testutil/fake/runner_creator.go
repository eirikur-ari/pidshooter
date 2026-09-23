package fake

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

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

// ErrHandler returns a Handler that discards everything it logs.
func (f *RunnerCreator) ErrHandler() *apperror.Handler {
	return apperror.NewHandler(&Logger{})
}
