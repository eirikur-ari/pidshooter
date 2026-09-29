package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// ConfigStore is a test double for outbound.ConfigStore.
type ConfigStore struct {
	Result  outbound.ConfigStoreResult
	LoadErr error
	SaveErr error
	Saved   *outbound.ConfigStoreResult // captured by the most recent Save call, nil if Save was never called
}

func (f *ConfigStore) Load() (outbound.ConfigStoreResult, error) {
	if f.LoadErr != nil {
		return outbound.ConfigStoreResult{}, f.LoadErr
	}
	return f.Result, nil
}

func (f *ConfigStore) Save(result outbound.ConfigStoreResult) error {
	f.Saved = &result
	return f.SaveErr
}
