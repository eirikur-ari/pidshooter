package game

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

// killResult is the result of one kill attempt on a target.
type killResult struct {
	target *game.Target
	err    error // the error from the kill attempt; nil when the kill succeeded
}

// applyTo applies the result to its target and the tracker.
func (r killResult) applyTo(tracker *killTracker) {
	r.target.CeaseFire()
	switch {
	case r.shouldReap():
		if r.target.Reap() {
			tracker.recordDud(r.target)
		}
	case r.err != nil:
		tracker.recordFailure(r.target, r.err)
	default:
		if r.target.Kill() {
			tracker.recordKill(r.target.Info.Rss)
		}
	}
}

// shouldReap reports whether the kill found its process already gone.
func (r killResult) shouldReap() bool {
	return apperror.CodeProcessNotFound.In(r.err)
}
