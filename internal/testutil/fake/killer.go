package fake

// Killer is a test double for gamedriven.ProcessKiller.
type Killer struct {
	KilledPIDs  []int
	KilledNames []string
	Err         error
}

func (f *Killer) Kill(pid int, name string) error {
	f.KilledPIDs = append(f.KilledPIDs, pid)
	f.KilledNames = append(f.KilledNames, name)
	return f.Err
}