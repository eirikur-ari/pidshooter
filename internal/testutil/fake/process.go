package fake

import "github.com/eirikur-ari/pidshooter/internal/core/process"

// Process is a test double for outbound.Process.
type Process struct {
	Processes   []process.Info
	FindErr     error
	KilledPIDs  []int
	KilledNames []string
	KillErr     error
}

func (f *Process) List() ([]process.Info, error)           { return f.Processes, f.FindErr }
func (f *Process) Find(_ []string) ([]process.Info, error) { return f.Processes, f.FindErr }

func (f *Process) Kill(pid int, name string) (bool, error) {
	f.KilledPIDs = append(f.KilledPIDs, pid)
	f.KilledNames = append(f.KilledNames, name)
	return f.KillErr == nil, f.KillErr
}
