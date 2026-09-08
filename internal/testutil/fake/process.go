package fake

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// Process is a test double for outbound.ProcessManager.
type Process struct {
	Infos           []outbound.ProcessInfo
	ListErr         error
	OwnPIDValue     int
	OwnUIDValue     int
	LookupNameValue string
	LookupNameErr   error
	KilledPIDs      []int
	KillErr         error
}

func (f *Process) List() ([]outbound.ProcessInfo, error) { return f.Infos, f.ListErr }
func (f *Process) OwnPID() int                           { return f.OwnPIDValue }
func (f *Process) OwnUID() int                           { return f.OwnUIDValue }
func (f *Process) LookupName(_ int) (string, error)      { return f.LookupNameValue, f.LookupNameErr }

func (f *Process) Kill(pid int) (bool, error) {
	f.KilledPIDs = append(f.KilledPIDs, pid)
	return f.KillErr == nil, f.KillErr
}
