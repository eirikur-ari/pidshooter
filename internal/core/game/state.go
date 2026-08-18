package game

import "sync/atomic"

// lifecycle represents the progression of a game session from creation to completion.
type lifecycle int32

const (
	pending lifecycle = iota // not yet started
	running                  // game loop active
	stopped                  // game over
)

// state is an atomic wrapper for lifecycle, safe for concurrent access.
type state struct {
	value atomic.Int32
}

// newState returns a state initialized to pending.
func newState() state { return state{} }

// Store sets the current lifecycle.
func (s *state) Store(l lifecycle) {
	s.value.Store(int32(l))
}

// Load returns the current lifecycle.
func (s *state) Load() lifecycle {
	return lifecycle(s.value.Load())
}
