package fake

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// Process is a test double for outbound.Process.
type Process struct {
	Processes   []outbound.ProcessInfo
	FindErr     error
	KilledPIDs  []int
	KilledNames []string
	KillErr     error
}

func (f *Process) List() ([]outbound.ProcessInfo, error)           { return f.Processes, f.FindErr }
func (f *Process) Find(_ []string) ([]outbound.ProcessInfo, error) { return f.Processes, f.FindErr }

func (f *Process) Kill(pid int, name string) (bool, error) {
	f.KilledPIDs = append(f.KilledPIDs, pid)
	f.KilledNames = append(f.KilledNames, name)
	return f.KillErr == nil, f.KillErr
}
