package game

import (
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// roster owns the set of targets in play for a session.
type roster struct {
	processes []process.Info
	targets   []*Target
}

// newRoster returns an empty roster; call spawn once bounds are known to populate it.
func newRoster(processes []process.Info) roster {
	return roster{processes: processes, targets: make([]*Target, 0, len(processes))}
}

// spawn creates a target for each process at a random position within bounds.
func (r *roster) spawn(bounds movement.Bounds) {
	for _, p := range r.processes {
		r.targets = append(r.targets, NewTarget(p, bounds))
	}
}

// move advances every target one step within bounds at the given speed.
func (r *roster) move(bounds movement.Bounds, speed float64) {
	for _, t := range r.targets {
		t.Update(bounds, speed)
	}
}

// allDead reports whether every target has finished its kill animation.
// No targets mean nothing was ever killed, not that everything was.
func (r *roster) allDead() bool {
	// An empty roster is not considered "all dead" because it means no targets were ever spawned.
	if r.isEmpty() {
		return false
	}
	for _, t := range r.targets {
		if !t.isDead() {
			return false
		}
	}
	return true
}

// isEmpty reports whether this roster holds no targets at all.
func (r *roster) isEmpty() bool { return len(r.targets) == 0 }

// hitAt returns the target whose tag covers the given game-space coordinates,
// or nil if none is hit.
func (r *roster) hitAt(x, y int) *Target {
	for _, t := range r.targets {
		if t.isHitAt(x, y) {
			return t
		}
	}
	return nil
}

// available returns the targets still in play — dead targets excluded —
// along with how many are currently alive.
func (r *roster) available() ([]*Target, int) {
	targets := make([]*Target, 0, len(r.targets))
	alive := 0
	for _, t := range r.targets {
		if t.isDead() {
			continue
		}
		if t.isAlive() {
			alive++
		}
		targets = append(targets, t)
	}
	return targets, alive
}
