package runner

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
)

// MockConfigLoader is a test double for configLoader.
type MockConfigLoader struct{ mock.Mock }

func (m *MockConfigLoader) Load() (config.Result, error) {
	args := m.Called()
	return args.Get(0).(config.Result), args.Error(1)
}
