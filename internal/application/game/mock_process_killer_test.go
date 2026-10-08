package game

import "github.com/stretchr/testify/mock"

// MockProcessKiller is a test double for processKiller.
type MockProcessKiller struct{ mock.Mock }

func (m *MockProcessKiller) Kill(pid int, name string, protected bool) error {
	args := m.Called(pid, name, protected)
	return args.Error(0)
}
