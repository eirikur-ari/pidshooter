package testutil

// FakeKiller is a test double for gamedriven.ProcessKiller.
type FakeKiller struct {
	KilledPIDs []int
	Err        error
}

func (f *FakeKiller) Kill(pid int) error {
	f.KilledPIDs = append(f.KilledPIDs, pid)
	return f.Err
}
