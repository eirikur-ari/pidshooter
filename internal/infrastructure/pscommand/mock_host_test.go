package pscommand

import "github.com/stretchr/testify/mock"

type MockHost struct{ mock.Mock }

func (m *MockHost) run(args ...string) ([]byte, error) {
	arguments := m.Called(args)
	data, _ := arguments.Get(0).([]byte)
	return data, arguments.Error(1)
}

func (m *MockHost) find(pid int) (osProcess, error) {
	arguments := m.Called(pid)
	process, _ := arguments.Get(0).(osProcess)
	return process, arguments.Error(1)
}

func (m *MockHost) pid() int {
	return m.Called().Int(0)
}

func (m *MockHost) uid() int {
	return m.Called().Int(0)
}

func (m *MockHost) goos() string {
	return m.Called().String(0)
}
