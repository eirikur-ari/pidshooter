package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// ConfigStore is a test double for outbound.ConfigStore.
type ConfigStore struct {
	Config  outbound.Config
	LoadErr error
	SaveErr error
	Saved   *outbound.Config // captured by the most recent Save call, nil if Save was never called
}

func (f *ConfigStore) Load() (outbound.Config, error) {
	if f.LoadErr != nil {
		return outbound.Config{}, f.LoadErr
	}
	return f.Config, nil
}

func (f *ConfigStore) Save(config outbound.Config) error {
	f.Saved = &config
	return f.SaveErr
}
