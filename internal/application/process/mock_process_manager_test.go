package process

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

type MockProcessManager struct{ mock.Mock }

func (m *MockProcessManager) Discover() ([]outbound.ProcessInfo, error) {
	args := m.Called()
	infos, _ := args.Get(0).([]outbound.ProcessInfo)
	return infos, args.Error(1)
}

func (m *MockProcessManager) OwnPID() int {
	return m.Called().Int(0)
}

func (m *MockProcessManager) OwnUID() int {
	return m.Called().Int(0)
}

func (m *MockProcessManager) LookupName(pid int) (string, error) {
	args := m.Called(pid)
	return args.String(0), args.Error(1)
}

func (m *MockProcessManager) Pin(pid int) (outbound.ProcessHandle, error) {
	args := m.Called(pid)
	handle, _ := args.Get(0).(outbound.ProcessHandle)
	return handle, args.Error(1)
}
