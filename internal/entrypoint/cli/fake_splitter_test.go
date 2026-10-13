package cli

type FakeSplitter struct {
	Patterns []string
	FlagArgs []string
	Err      error
}

func (f *FakeSplitter) split([]string) (patterns, flagArgs []string, err error) {
	return f.Patterns, f.FlagArgs, f.Err
}
