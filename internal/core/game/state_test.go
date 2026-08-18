package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateDefaultIsPending(t *testing.T) {
	var s state
	assert.Equal(t, pending, s.Load(), "expected zero-value currentState to be Pending")
}

func TestStateStoreLoad(t *testing.T) {
	var s state
	s.Store(running)
	assert.Equal(t, running, s.Load(), "expected Running after Store")
	s.Store(stopped)
	assert.Equal(t, stopped, s.Load(), "expected Stopped after Store")
}
