package input

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

func TestDispatcher_Dispatch_PassesItsInputToEventAndReturnsEventResult(t *testing.T) {
	input := newInputFixture(newGameSessionFixture(nil, game.Config{}))
	expected := &game.Target{}
	event := &FakeEvent{Result: expected}

	result := NewDispatcher(input).Dispatch(event)

	assert.Same(t, input, event.Received)
	assert.Same(t, expected, result)
}
