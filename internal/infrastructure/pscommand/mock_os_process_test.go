package pscommand

import (
	"os"

	"github.com/stretchr/testify/mock"
)

type MockOsProcess struct{ mock.Mock }

func (m *MockOsProcess) Signal(signal os.Signal) error {
	return m.Called(signal).Error(0)
}

func (m *MockOsProcess) Release() error {
	return m.Called().Error(0)
}
