package config

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

func configStoreMockSetup(stored outbound.Config, loadErr error) *MockConfigStore {
	store := &MockConfigStore{}
	store.On("Load").Return(stored, loadErr)
	return store
}
