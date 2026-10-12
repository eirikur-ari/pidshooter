package process

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

func discoveryMockSetup(infos []outbound.ProcessInfo, ownPID, ownUID int, discoverErr error) *MockProcessManager {
	manager := &MockProcessManager{}
	manager.On("Discover").Return(infos, discoverErr)
	manager.On("OwnPID").Return(ownPID)
	manager.On("OwnUID").Return(ownUID)
	return manager
}

func pinnedProcessMockSetup(pid int, pinErr error, name string, lookupErr, killErr error) (*MockProcessManager, *MockProcessHandle) {
	handle := &MockProcessHandle{}
	handle.On("Kill").Return(killErr)
	handle.On("Release").Return(nil)

	manager := &MockProcessManager{}
	if pinErr != nil {
		manager.On("Pin", pid).Return(nil, pinErr)
	} else {
		manager.On("Pin", pid).Return(handle, nil)
	}
	manager.On("LookupName", pid).Return(name, lookupErr)
	return manager, handle
}
