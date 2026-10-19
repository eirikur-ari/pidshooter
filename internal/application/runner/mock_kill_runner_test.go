package runner

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

type MockKillRunner struct{ mock.Mock }

func (m *MockKillRunner) Run(request inbound.KillRequest) error {
	return m.Called(request).Error(0)
}
