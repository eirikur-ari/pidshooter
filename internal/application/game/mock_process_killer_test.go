package game

import "github.com/stretchr/testify/mock"

type MockProcessKiller struct{ mock.Mock }

func (m *MockProcessKiller) Kill(pid int, name string, protected bool) error {
	args := m.Called(pid, name, protected)
	return args.Error(0)
}
