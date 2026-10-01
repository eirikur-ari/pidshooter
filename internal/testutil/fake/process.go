package fake

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// ProcessManager is a test double for outbound.ProcessManager.
type ProcessManager struct {
	Infos           []outbound.ProcessInfo
	DiscoverErr     error
	OwnPIDValue     int
	OwnUIDValue     int
	LookupNameValue string
	LookupNameErr   error
	PinErr          error
	KilledPIDs      []int
	KillErr         error
	ReleasedPIDs    []int
}

func (f *ProcessManager) Discover() ([]outbound.ProcessInfo, error) { return f.Infos, f.DiscoverErr }
func (f *ProcessManager) OwnPID() int                               { return f.OwnPIDValue }
func (f *ProcessManager) OwnUID() int                               { return f.OwnUIDValue }
func (f *ProcessManager) LookupName(_ int) (string, error)          { return f.LookupNameValue, f.LookupNameErr }

func (f *ProcessManager) Pin(pid int) (outbound.ProcessHandle, error) {
	if f.PinErr != nil {
		return nil, f.PinErr
	}
	return &processHandle{process: f, pid: pid}, nil
}
