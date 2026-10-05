package cli

// FakeRunner is a test double for inbound.Runner. It returns Err.
type FakeRunner struct {
	Err error
	// Calls counts how many times Run was called.
	Calls int
}

// Run records the call and returns Err.
func (r *FakeRunner) Run() error {
	r.Calls++
	return r.Err
}
