package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicLifecycleDefaultIsPending(t *testing.T) {
	var s atomicLifecycle
	assert.Equal(t, pending, s.Load(), "expected zero-value currentState to be Pending")
}

func TestAtomicLifecycleStoreLoad(t *testing.T) {
	var s atomicLifecycle
	s.Store(running)
	assert.Equal(t, running, s.Load(), "expected Running after Store")
	s.Store(stopped)
	assert.Equal(t, stopped, s.Load(), "expected Stopped after Store")
}
