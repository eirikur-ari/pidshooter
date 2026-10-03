package testutil

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// FakeConfigStore is a test double for outbound.ConfigStore.
type FakeConfigStore struct {
	Config  outbound.Config
	LoadErr error
	SaveErr error
	Saved   *outbound.Config // captured by the most recent Save call, nil if Save was never called
}

func (f *FakeConfigStore) Load() (outbound.Config, error) {
	if f.LoadErr != nil {
		return outbound.Config{}, f.LoadErr
	}
	return f.Config, nil
}

func (f *FakeConfigStore) Save(config outbound.Config) error {
	f.Saved = &config
	return f.SaveErr
}
