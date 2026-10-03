package testutil

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// FakeProcessManager is a test double for outbound.ProcessManager.
type FakeProcessManager struct {
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

func (f *FakeProcessManager) Discover() ([]outbound.ProcessInfo, error) {
	return f.Infos, f.DiscoverErr
}
func (f *FakeProcessManager) OwnPID() int { return f.OwnPIDValue }
func (f *FakeProcessManager) OwnUID() int { return f.OwnUIDValue }
func (f *FakeProcessManager) LookupName(_ int) (string, error) {
	return f.LookupNameValue, f.LookupNameErr
}

func (f *FakeProcessManager) Pin(pid int) (outbound.ProcessHandle, error) {
	if f.PinErr != nil {
		return nil, f.PinErr
	}
	return &fakeProcessHandle{process: f, pid: pid}, nil
}
