package game

import "sync/atomic"

// lifecycle represents the progression of a game session from creation to completion.
type lifecycle int32

const (
	pending lifecycle = iota // not yet started
	running                  // game loop active
	stopped                  // game over
)

// atomicLifecycle is an atomic wrapper for lifecycle, safe for concurrent access.
type atomicLifecycle struct {
	value atomic.Int32
}

// newAtomicLifecycle returns an atomicLifecycle initialized to pending.
func newAtomicLifecycle() atomicLifecycle { return atomicLifecycle{} }

// Store sets the current lifecycle.
func (s *atomicLifecycle) Store(l lifecycle) {
	s.value.Store(int32(l))
}

// Load returns the current lifecycle.
func (s *atomicLifecycle) Load() lifecycle {
	return lifecycle(s.value.Load())
}
