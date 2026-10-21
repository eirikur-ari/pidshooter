package composition

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

type MockPlayRunner struct{ mock.Mock }

func (m *MockPlayRunner) Run(request inbound.PlayRequest) error {
	return m.Called(request).Error(0)
}
