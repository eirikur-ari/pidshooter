package config

import (
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func newStoredConfigFixture() outbound.Config {
	return outbound.Config{
		Process: outbound.ProcessConfig{IncludeRoot: testutil.Pointer(true)},
		Game: outbound.GameConfig{
			ConfirmMode: testutil.Pointer(true),
			Speed:       testutil.Pointer(3.0),
			TimeLimit:   testutil.Pointer(45),
		},
	}
}

func newOptionsFixture() Options {
	return Options{
		Process: ProcessOptions{IncludeRoot: testutil.Pointer(true)},
		Game: GameOptions{
			ConfirmMode: testutil.Pointer(true),
			Speed:       testutil.Pointer(5.0),
			TimeLimit:   testutil.Pointer(60),
		},
	}
}
