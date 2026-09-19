package fake

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// Process is a test double for outbound.Process.
type Process struct {
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

func (f *Process) Discover() ([]outbound.ProcessInfo, error) { return f.Infos, f.DiscoverErr }
func (f *Process) OwnPID() int                               { return f.OwnPIDValue }
func (f *Process) OwnUID() int                               { return f.OwnUIDValue }
func (f *Process) LookupName(_ int) (string, error)          { return f.LookupNameValue, f.LookupNameErr }

func (f *Process) Pin(pid int) (outbound.ProcessHandle, error) {
	if f.PinErr != nil {
		return nil, f.PinErr
	}
	return &processHandle{process: f, pid: pid}, nil
}
