package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState_DefaultIsPending(t *testing.T) {
	var s state
	assert.Equal(t, Pending, s.Load(), "expected zero-value State to be Pending")
}

func TestState_Store_Load(t *testing.T) {
	var s state
	s.Store(Running)
	assert.Equal(t, Running, s.Load(), "expected Running after Store")
	s.Store(Stopped)
	assert.Equal(t, Stopped, s.Load(), "expected Stopped after Store")
}

func TestState_CompareAndSwap_Success(t *testing.T) {
	var s state
	assert.True(t, s.CompareAndSwap(Pending, Running), "expected CAS to succeed when old matches current")
	assert.Equal(t, Running, s.Load(), "expected Running after successful CAS")
}

func TestState_CompareAndSwap_Failure(t *testing.T) {
	var s state
	s.Store(Running)
	assert.False(t, s.CompareAndSwap(Pending, Stopped), "expected CAS to fail when old does not match current")
	assert.Equal(t, Running, s.Load(), "expected state unchanged after failed CAS")
}
