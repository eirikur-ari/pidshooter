package runner

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

type MockListRunner struct{ mock.Mock }

func (m *MockListRunner) Run(request inbound.ListRequest) error {
	return m.Called(request).Error(0)
}
