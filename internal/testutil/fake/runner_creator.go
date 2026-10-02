package fake

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// RunnerCreator is a test double that constructs a Runner.
type RunnerCreator struct {
	Runner inbound.Runner
	Err    error
	// Calls counts how many times Create was called.
	Calls int
	// Patterns and Options capture the arguments Create was most recently called with.
	Patterns []string
	Options  config.Options
}

// Create returns Runner and Err, and records the call, patterns, and config.Options.
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
