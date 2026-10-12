package process

import "github.com/stretchr/testify/mock"

type MockProcessHandle struct{ mock.Mock }

func (m *MockProcessHandle) Kill() error {
	return m.Called().Error(0)
}

func (m *MockProcessHandle) Release() error {
	return m.Called().Error(0)
}
