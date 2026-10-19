package cli

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

type MockRunner struct{ mock.Mock }

func (m *MockRunner) Run(request inbound.RunRequest) error {
	return m.Called(request).Error(0)
}

func (m *MockRunner) HandleError(err error) error {
	result := m.Called(err).Get(0)
	if handle, ok := result.(func(error) error); ok {
		return handle(err)
	}
	if result == nil {
		return nil
	}
	return result.(error)
}
