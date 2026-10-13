package config

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

type MockConfigStore struct{ mock.Mock }

func (m *MockConfigStore) Load() (outbound.Config, error) {
	args := m.Called()
	config, _ := args.Get(0).(outbound.Config)
	return config, args.Error(1)
}

func (m *MockConfigStore) Save(config outbound.Config) error {
	return m.Called(config).Error(0)
}
