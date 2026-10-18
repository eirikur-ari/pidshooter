package config

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// toConfig returns the Result as a stored config, without the fields that are
// never stored.
func (result Result) toConfig() outbound.Config {
	return outbound.Config{
		Process: outbound.ProcessConfig{IncludeRoot: &result.Process.IncludeRoot},
		Game: outbound.GameConfig{
			ConfirmMode: &result.Game.ConfirmMode,
			Speed:       &result.Game.Speed,
			TimeLimit:   &result.Game.TimeLimit,
		},
	}
}
