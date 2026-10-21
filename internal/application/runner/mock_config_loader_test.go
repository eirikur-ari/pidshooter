package runner

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
)

type MockConfigLoader struct{ mock.Mock }

func (m *MockConfigLoader) Load(options config.Options) (config.Result, error) {
	args := m.Called(options)
	return args.Get(0).(config.Result), args.Error(1)
}
