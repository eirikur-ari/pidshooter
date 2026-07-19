package fake

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// Process is a test double for outbound.Process.
type Process struct {
	Processes       []outbound.ProcessInfo
	ListErr         error
	OwnPidValue     int
	LookupNameValue string
	LookupNameErr   error
	KilledPIDs      []int
	KillErr         error
}

func (f *Process) List() ([]outbound.ProcessInfo, error) { return f.Processes, f.ListErr }
func (f *Process) OwnPid() int                           { return f.OwnPidValue }
func (f *Process) LookupName(_ int) (string, error)      { return f.LookupNameValue, f.LookupNameErr }

func (f *Process) Kill(pid int) (bool, error) {
	f.KilledPIDs = append(f.KilledPIDs, pid)
	return f.KillErr == nil, f.KillErr
}
