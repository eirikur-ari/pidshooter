package cli

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

// FakeRunnerCreator is a test double that constructs a Runner.
type FakeRunnerCreator struct {
	Runner inbound.Runner
	Err    error
	// Calls counts how many times Create was called.
	Calls int
	// Patterns and Options capture the arguments Create was most recently
	// called with.
	Patterns []string
	Options  config.Options
}

// Create returns Runner and Err, and records the call, patterns, and
// config.Options.
func (f *FakeRunnerCreator) Create(patterns []string, options config.Options) (inbound.Runner, error) {
	f.Calls++
	f.Patterns = patterns
	f.Options = options
	return f.Runner, f.Err
}

// ErrHandler returns a Handler that discards everything it logs.
func (f *FakeRunnerCreator) ErrHandler() *apperror.Handler {
	return apperror.NewHandler(&testutil.FakeLogger{})
}
