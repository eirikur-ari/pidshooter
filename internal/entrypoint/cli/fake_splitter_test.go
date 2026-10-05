package cli

// FakeSplitter is a test double for argumentSplitter. It returns Patterns,
// FlagArgs and Err.
type FakeSplitter struct {
	Patterns []string
	FlagArgs []string
	Err      error
}

// split returns Patterns, FlagArgs and Err.
func (f *FakeSplitter) split([]string) (patterns, flagArgs []string, err error) {
	return f.Patterns, f.FlagArgs, f.Err
}
