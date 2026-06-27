package fake

// Killer is a test double for gamedriven.ProcessKiller.
type Killer struct {
	KilledPIDs []int
	Err        error
}

func (f *Killer) Kill(pid int) error {
	f.KilledPIDs = append(f.KilledPIDs, pid)
	return f.Err
}