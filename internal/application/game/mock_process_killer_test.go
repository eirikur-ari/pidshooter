package game

import "github.com/stretchr/testify/mock"

// MockProcessKiller is a test double for processKiller.
type MockProcessKiller struct{ mock.Mock }

func (m *MockProcessKiller) Kill(pid int, name string, protected bool) (shouldReap bool, err error) {
	args := m.Called(pid, name, protected)
	return args.Bool(0), args.Error(1)
}
