package config

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func newStoredConfigFixture() outbound.Config {
	return outbound.Config{
		Process: outbound.ProcessConfig{IncludeRoot: new(true)},
		Game: outbound.GameConfig{
			ConfirmMode: new(true),
			Speed:       new(3.0),
			TimeLimit:   new(45),
		},
	}
}

func newOptionsFixture() Options {
	return Options{
		Process: ProcessOptions{IncludeRoot: new(true)},
		Game: GameOptions{
			ConfirmMode: new(true),
			Speed:       new(5.0),
			TimeLimit:   new(60),
		},
	}
}
