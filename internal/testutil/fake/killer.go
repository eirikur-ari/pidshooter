package fake

import "github.com/eirikur-ari/pidshooter/internal/core/game"

// Killer adapts a plain function to the game.Service killer interface for tests.
type Killer func(target *game.Target) (killed, shouldReap bool, err error)

func (f Killer) Kill(target *game.Target) (killed, shouldReap bool, err error) {
	return f(target)
}
