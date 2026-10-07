package input

import "github.com/eirikur-ari/pidshooter/internal/core/game"

type FakeEvent struct {
	Received *game.Input
	Result   *game.Target
}

func (e *FakeEvent) dispatch(input *game.Input) *game.Target {
	e.Received = input
	return e.Result
}
