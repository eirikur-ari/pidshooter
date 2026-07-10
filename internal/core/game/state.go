package game

import "sync/atomic"

// Lifecycle represents the progression of a game session from creation to completion.
type Lifecycle int32

const (
	Pending Lifecycle = iota // not yet started
	Running                  // game loop active
	Stopped                  // game over
)

// state is an atomic wrapper for Lifecycle, safe for concurrent access.
type state struct {
	value atomic.Int32
}

// Store sets the current lifecycle.
func (s *state) Store(l Lifecycle) {
	s.value.Store(int32(l))
}

// Load returns the current lifecycle.
func (s *state) Load() Lifecycle {
	return Lifecycle(s.value.Load())
}

// CompareAndSwap atomically transitions from old to new, reporting whether it succeeded.
func (s *state) CompareAndSwap(old, new Lifecycle) bool {
	return s.value.CompareAndSwap(int32(old), int32(new))
}
