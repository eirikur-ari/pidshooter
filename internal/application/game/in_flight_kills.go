package game

import (
	"sync"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/core/game"
)

// processKiller kills the process behind a target. shouldReap is true when
// that process was already gone, so the target should be reaped instead.
type processKiller interface {
	Kill(pid int, name string, protected bool) (shouldReap bool, err error)
}

// killResult is the result of one kill attempt on a target.
type killResult struct {
	target     *game.Target
	shouldReap bool
	err        error // non-nil when the kill failed and the target must stay alive
}

// inFlightKills runs kills in the background and applies their results to a tracker.
type inFlightKills struct {
	killer      processKiller
	gracePeriod time.Duration
	tracker     *killTracker
	results     chan killResult // channel buffer capacity = 10 kills in flight
	waitGroup   sync.WaitGroup
	done        chan struct{}
}

// newInFlightKills returns an inFlightKills that kills through killer and
// records results in tracker, waiting at most gracePeriod for kills still
// in progress once the session has ended.
func newInFlightKills(killer processKiller, gracePeriod time.Duration, tracker *killTracker) *inFlightKills {
	return &inFlightKills{
		killer:      killer,
		gracePeriod: gracePeriod,
		tracker:     tracker,
		results:     make(chan killResult, 10),
		done:        make(chan struct{}),
	}
}

// start fires at target and kills its process in the background.
func (k *inFlightKills) start(target *game.Target) {
	target.FireShot()
	k.waitGroup.Go(func() { k.killOrReap(target) })
}

// killOrReap kills target and reports the result, unless the kills have been closed first.
func (k *inFlightKills) killOrReap(target *game.Target) {
	shouldReap, err := k.killer.Kill(target.Info.PID, target.Info.Name, target.Info.IsProtected())

	result := killResult{target: target}
	switch {
	case shouldReap:
		result.shouldReap = true
	case err != nil:
		result.err = err
	}

	select {
	case k.results <- result:
	case <-k.done:
	}
}

// close releases kills still waiting to report their result. Call it after
// awaitRemaining, since results reported after close may be dropped.
func (k *inFlightKills) close() {
	close(k.done)
}

// awaitRemaining applies the results of kills still in progress, waiting at
// most the grace period. Later results are dropped.
func (k *inFlightKills) awaitRemaining() {
	finished := make(chan struct{})
	go func() {
		k.waitGroup.Wait()
		close(finished)
	}()

	timer := time.NewTimer(k.gracePeriod)
	defer timer.Stop()

	for {
		select {
		case result := <-k.results:
			result.applyTo(k.tracker)
		case <-finished:
			k.applyFinished()
			return
		case <-timer.C:
			k.applyFinished()
			return
		}
	}
}

// applyFinished applies the result of every kill that has finished, without blocking.
func (k *inFlightKills) applyFinished() {
	for {
		select {
		case result := <-k.results:
			result.applyTo(k.tracker)
		default:
			return
		}
	}
}

// applyTo applies the result to its target and the tracker.
func (r killResult) applyTo(tracker *killTracker) {
	r.target.CeaseFire()
	switch {
	case r.err != nil:
		tracker.recordFailure(r.target, r.err)
	case r.shouldReap:
		if r.target.Reap() {
			tracker.recordDud(r.target)
		}
	default:
		if r.target.Kill() {
			tracker.recordKill(r.target.Info.Rss)
		}
	}
}
