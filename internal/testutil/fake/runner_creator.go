package fake

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// RunnerCreator is a test double that constructs a Runner. It returns
// Runner and Err, and records how many times Create was called and the
// patterns and config.Options it was last called with.
type RunnerCreator struct {
	Runner   inbound.Runner
	Err      error
	Calls    int
	Patterns []string
	Options  config.Options
}

// Create returns Runner and Err, and records the call, patterns, and opts.
func (f *RunnerCreator) Create(patterns []string, opts config.Options) (inbound.Runner, error) {
	f.Calls++
	f.Patterns = patterns
	f.Options = opts
	return f.Runner, f.Err
}

// ErrHandler returns a Handler that discards everything it logs.
func (f *RunnerCreator) ErrHandler() *apperror.Handler {
	return apperror.NewHandler(&Logger{})
}
